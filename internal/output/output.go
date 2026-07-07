package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/itchyny/gojq"
)

type Format string

const (
	FormatTable Format = "table"
	FormatText  Format = "text"
	FormatJSON  Format = "json"
)

type Options struct {
	JSONFields []string `json:"jsonFields,omitempty"`
	JQ         string   `json:"jq,omitempty"`
	Template   string   `json:"template,omitempty"`
	Plain      bool     `json:"plain"`
}

type Row []string

type Table struct {
	Headers []string `json:"headers"`
	Rows    []Row    `json:"rows"`
	Empty   string   `json:"empty"`
}

type DetailField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Detail struct {
	Title       string        `json:"title"`
	Fields      []DetailField `json:"fields"`
	BodyHeading string        `json:"bodyHeading,omitempty"`
	Body        string        `json:"body,omitempty"`
}

type Renderer struct {
	out io.Writer
}

func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{out: out}
}

func (r *Renderer) Render(value any, table Table, options Options) error {
	if len(options.JSONFields) > 0 {
		return r.renderJSON(value, options)
	}
	if options.Template != "" {
		return r.renderTemplate(value, options.Template)
	}
	return r.renderTable(table)
}

func (r *Renderer) RenderDetail(value any, detail Detail, options Options) error {
	if len(options.JSONFields) > 0 {
		return r.renderJSON(value, options)
	}
	if options.Template != "" {
		return r.renderTemplate(value, options.Template)
	}
	return r.renderDetail(detail)
}

func (r *Renderer) RenderText(value any, text string, options Options) error {
	if len(options.JSONFields) > 0 {
		return r.renderJSON(value, options)
	}
	if options.Template != "" {
		return r.renderTemplate(value, options.Template)
	}
	_, err := fmt.Fprint(r.out, text)
	return err
}

func (r *Renderer) renderTable(table Table) error {
	if len(table.Rows) == 0 {
		if table.Empty != "" {
			_, err := fmt.Fprintln(r.out, table.Empty)
			return err
		}
		return nil
	}
	writer := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	if len(table.Headers) > 0 {
		fmt.Fprintln(writer, strings.Join(table.Headers, "\t"))
	}
	for _, row := range table.Rows {
		fmt.Fprintln(writer, strings.Join([]string(row), "\t"))
	}
	return writer.Flush()
}

func (r *Renderer) renderDetail(detail Detail) error {
	if detail.Title != "" {
		fmt.Fprintln(r.out, detail.Title)
	}
	for _, field := range detail.Fields {
		fmt.Fprintf(r.out, "  - %s: %s\n", field.Name, field.Value)
	}
	if detail.Body != "" {
		if detail.BodyHeading != "" {
			fmt.Fprintf(r.out, "\n%s:\n", detail.BodyHeading)
		} else {
			fmt.Fprintln(r.out)
		}
		fmt.Fprintln(r.out, detail.Body)
	}
	return nil
}

func (r *Renderer) renderTemplate(value any, rawTemplate string) error {
	tmpl, err := template.New("output").Parse(rawTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(r.out, value)
}

func (r *Renderer) renderJSON(value any, options Options) error {
	filtered, err := SelectFields(value, options.JSONFields)
	if err != nil {
		return err
	}
	if options.JQ != "" {
		filtered, err = ApplyJQ(filtered, options.JQ)
		if err != nil {
			return err
		}
	}
	encoder := json.NewEncoder(r.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(filtered)
}

func SelectFields(value any, fields []string) (any, error) {
	normalized := normalize(value)
	if len(fields) == 0 {
		return normalized, nil
	}
	if items, ok := normalized.([]any); ok {
		selected := make([]any, 0, len(items))
		for _, item := range items {
			selectedItem, err := selectObjectFields(item, fields)
			if err != nil {
				return nil, err
			}
			selected = append(selected, selectedItem)
		}
		return selected, nil
	}
	return selectObjectFields(normalized, fields)
}

func selectObjectFields(value any, fields []string) (map[string]any, error) {
	selected := make(map[string]any, len(fields))
	source, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("--json fields require object output")
	}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		value, ok := source[field]
		if !ok {
			return nil, fmt.Errorf("unknown JSON field %q; available fields: %s", field, strings.Join(sortedKeys(source), ", "))
		}
		selected[field] = value
	}
	return selected, nil
}

func ApplyJQ(value any, expression string) (any, error) {
	query, err := gojq.Parse(expression)
	if err != nil {
		return nil, err
	}
	iter := query.Run(value)
	results := make([]any, 0)
	for {
		next, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := next.(error); ok {
			return nil, err
		}
		results = append(results, next)
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

func normalize(value any) any {
	body, err := json.Marshal(value)
	if err != nil {
		return value
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return value
	}
	return decoded
}

func sortedKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func Fields(value any) []string {
	reflected := reflect.TypeOf(value)
	if reflected.Kind() == reflect.Pointer {
		reflected = reflected.Elem()
	}
	if reflected.Kind() != reflect.Struct {
		return nil
	}
	fields := make([]string, 0, reflected.NumField())
	for index := 0; index < reflected.NumField(); index++ {
		field := reflected.Field(index)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		fields = append(fields, name)
	}
	return fields
}

func Capture(fn func(*Renderer) error) (string, error) {
	buffer := bytes.Buffer{}
	if err := fn(NewRenderer(&buffer)); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func ParseJSONFields(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			fields = append(fields, part)
		}
	}
	return fields
}

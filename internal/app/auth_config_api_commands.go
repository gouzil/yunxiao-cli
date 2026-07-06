package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gouzil/yunxiao-cli/internal/api"
	"github.com/gouzil/yunxiao-cli/internal/auth"
	"github.com/gouzil/yunxiao-cli/internal/config"
	"github.com/gouzil/yunxiao-cli/internal/output"
	"github.com/gouzil/yunxiao-cli/internal/yunxiao"
	"github.com/spf13/cobra"
)

func (r *Root) newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Yunxiao",
	}
	cmd.AddCommand(r.newAuthLoginCommand(), r.newAuthStatusCommand(), r.newAuthLogoutCommand())
	return cmd
}

func (r *Root) newAuthLoginCommand() *cobra.Command {
	var withToken bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in using a Yunxiao personal access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := r.flags.Endpoint
			if endpoint == "" {
				endpoint = "openapi-rdc.aliyuncs.com"
			}
			var token string
			if withToken {
				body, err := io.ReadAll(r.io.In)
				if err != nil {
					return err
				}
				token = strings.TrimSpace(string(body))
				if token == "" {
					return fmt.Errorf("token is required on stdin when using --with-token")
				}
			} else {
				fmt.Fprintf(r.io.Out, "Yunxiao endpoint [%s]: ", endpoint)
				reader := bufio.NewReader(r.io.In)
				line, err := reader.ReadString('\n')
				if err == nil && strings.TrimSpace(line) != "" {
					endpoint = strings.TrimSpace(line)
				}
				fmt.Fprintln(r.io.Out, "Create a Yunxiao personal access token: https://account-devops.aliyun.com/settings/personalAccessToken")
				fmt.Fprintln(r.io.Out, "Optional account lookup: 组织管理 / 用户 / 只读")
				fmt.Fprintln(r.io.Out, "Common CLI permissions:")
				fmt.Fprintln(r.io.Out, "  - 代码管理 / 用户资源、代码仓库、分支、提交、文件 / 只读; 合并请求 / 读写")
				fmt.Fprintln(r.io.Out, "  - 按需: SSH 密钥 / 读写; 流水线 / 流水线 / 只读; 流水线运行实例、流水线运行任务 / 读写")
				fmt.Fprintln(r.io.Out, "  - 按需: 项目协作 / 项目、项目成员、标签 / 只读; 工作项 / 读写")
				fmt.Fprint(r.io.Out, "Paste your Yunxiao personal access token: ")
				tokenLine, err := reader.ReadString('\n')
				if err != nil && !errors.Is(err, io.EOF) {
					return err
				}
				token = strings.TrimSpace(tokenLine)
			}
			fmt.Fprintf(r.io.Out, "- Saving token for %s...\n", endpoint)
			result, err := r.authManager.Login(cmd.Context(), auth.LoginRequest{Endpoint: endpoint, Token: token})
			if err != nil {
				return fmt.Errorf("Authentication failed for %s\n  - Message: %w", endpoint, err)
			}
			return r.renderer.RenderDetail(result, output.LoginSuccess(result), r.renderOptions())
		},
	}
	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read token from standard input")
	return cmd
}

func (r *Root) newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			credential, err := r.authManager.Status(resolved.Endpoint.Value)
			if err != nil {
				var notLogged auth.ErrNotLoggedIn
				if errors.As(err, &notLogged) {
					return r.renderer.RenderDetail(credential, output.AuthStatus(resolved.Endpoint.Value, credential, false), r.renderOptions())
				}
				return err
			}
			return r.renderer.RenderDetail(credential, output.AuthStatus(resolved.Endpoint.Value, credential, true), r.renderOptions())
		},
	}
}

func (r *Root) newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Delete local Yunxiao credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := r.resolvedConfig()
			if err != nil {
				return err
			}
			credential, err := r.authManager.Logout(resolved.Endpoint.Value)
			if err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Logged out of %s account %s\n", resolved.Endpoint.Value, credential.User.DisplayName)
			return nil
		},
	}
}

func (r *Root) newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "config", Short: "Manage local Yunxiao configuration"}
	cmd.AddCommand(r.newConfigListCommand(), r.newConfigGetCommand(), r.newConfigSetCommand())
	return cmd
}

func (r *Root) newConfigListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List effective configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := r.configStore.List(config.Values{
				Endpoint:     r.flags.Endpoint,
				Organization: r.flags.Organization,
				Project:      r.flags.Project,
				Repo:         r.flags.Repo,
			})
			if err != nil {
				return err
			}
			return r.renderer.Render(entries, output.ConfigTable(entries), r.renderOptions())
		},
	}
}

func (r *Root) newConfigGetCommand() *cobra.Command {
	var scopeText string
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  requireArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := config.ParseKey(args[0])
			if err != nil {
				return err
			}
			value := ""
			if cmd.Flags().Changed("scope") {
				scope, err := config.ParseScope(scopeText)
				if err != nil {
					return err
				}
				value, err = r.configStore.Get(scope, key)
				if err != nil {
					return err
				}
			} else {
				resolved, err := r.resolvedConfig()
				if err != nil {
					return err
				}
				value = resolvedConfigValue(resolved, key)
			}
			fmt.Fprintln(r.io.Out, value)
			return nil
		},
	}
	cmd.Flags().StringVar(&scopeText, "scope", string(config.ScopeGlobal), "Configuration scope: global or repo")
	return cmd
}

func resolvedConfigValue(resolved config.Resolved, key config.Key) string {
	switch key {
	case config.KeyEndpoint:
		return resolved.Endpoint.Value
	case config.KeyOrganization:
		return resolved.Organization.Value
	case config.KeyProject:
		return resolved.Project.Value
	case config.KeyRepo:
		return resolved.Repo.Value
	case config.KeyGitProtocol:
		return resolved.GitProtocol.Value
	default:
		return ""
	}
}

func (r *Root) newConfigSetCommand() *cobra.Command {
	var scopeText string
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := config.ParseKey(args[0])
			if err != nil {
				return err
			}
			scope, err := config.ParseScope(scopeText)
			if err != nil {
				return err
			}
			if err := r.configStore.Set(scope, key, args[1]); err != nil {
				return err
			}
			fmt.Fprintf(r.io.Out, "✓ Set %s to %s (%s)\n", key, args[1], scope)
			return nil
		},
	}
	cmd.Flags().StringVar(&scopeText, "scope", string(config.ScopeGlobal), "Configuration scope: global or repo")
	return cmd
}

func (r *Root) newAPICommand() *cobra.Command {
	var bodyFile string
	var bodyText string
	var headerValues []string
	cmd := &cobra.Command{
		Use:   "api <method> <path>",
		Short: "Call a Yunxiao OpenAPI path directly",
		Args:  requireArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readBody(bodyFile, bodyText)
			if err != nil {
				return err
			}
			headers, err := parseHeaders(headerValues)
			if err != nil {
				return err
			}
			result, err := r.services.RawAPI.Request(cmd.Context(), yunxiao.RawAPIRequest{
				Method:  strings.ToUpper(args[0]),
				Path:    args[1],
				Headers: headers,
				Body:    body,
				Verbose: r.flags.Verbose,
			})
			if err != nil {
				return err
			}
			if r.flags.Verbose {
				fmt.Fprintf(r.io.ErrOut, "> %s %s\n< HTTP %d\n", result.Method, args[1], result.Status)
				if result.RequestID != "" {
					fmt.Fprintf(r.io.ErrOut, "< x-acs-request-id: %s\n", result.RequestID)
				}
				fmt.Fprintln(r.io.ErrOut)
			}
			return r.renderAPIResult(result)
		},
	}
	cmd.Flags().StringVarP(&bodyFile, "input", "i", "", "Read request body from file")
	cmd.Flags().StringVarP(&bodyText, "body", "b", "", "Request body text")
	cmd.Flags().StringArrayVarP(&headerValues, "header", "H", nil, "Request header in Name: Value form")
	cmd.ValidArgs = []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete}
	return cmd
}

func (r *Root) renderAPIResult(result yunxiao.RawAPIResult) error {
	if r.flags.JSONFields != "" {
		return r.renderer.RenderText(result, result.Body, r.renderOptions())
	}
	if r.flags.JQ == "" && r.flags.Template == "" {
		_, err := fmt.Fprint(r.io.Out, result.Body)
		return err
	}
	value, jsonValue, err := decodeRawAPIBody(result.Body)
	if err != nil && r.flags.JQ != "" {
		return err
	}
	if r.flags.JQ != "" {
		filtered, err := output.ApplyJQ(value, r.flags.JQ)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(r.io.Out)
		return encoder.Encode(filtered)
	}
	if r.flags.Template != "" {
		target := any(result.Body)
		if jsonValue {
			target = value
		}
		tmpl, err := template.New("api").Parse(r.flags.Template)
		if err != nil {
			return err
		}
		return tmpl.Execute(r.io.Out, target)
	}
	return nil
}

func decodeRawAPIBody(body string) (any, bool, error) {
	var value any
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		return nil, false, fmt.Errorf("response body is not JSON: %w", err)
	}
	return value, true, nil
}

func readBody(path string, text string) ([]byte, error) {
	if path != "" && text != "" {
		return nil, fmt.Errorf("use only one of --input or --body")
	}
	if text != "" {
		return []byte(text), nil
	}
	if path == "" {
		return nil, nil
	}
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func parseHeaders(values []string) ([]api.Header, error) {
	headers := make([]api.Header, 0, len(values))
	for _, value := range values {
		name, headerValue, ok := strings.Cut(value, ":")
		if !ok {
			return nil, fmt.Errorf("header must be in Name: Value form")
		}
		headers = append(headers, api.Header{Name: strings.TrimSpace(name), Value: strings.TrimSpace(headerValue)})
	}
	return headers, nil
}

func readFileOrValue(path string, value string) (string, error) {
	if path != "" && value != "" {
		return "", fmt.Errorf("use only one of body value or body file")
	}
	if value != "" {
		return value, nil
	}
	if path == "" {
		return "", nil
	}
	if path == "-" {
		body, err := io.ReadAll(os.Stdin)
		return string(bytes.TrimSpace(body)), err
	}
	body, err := os.ReadFile(path)
	return string(bytes.TrimSpace(body)), err
}

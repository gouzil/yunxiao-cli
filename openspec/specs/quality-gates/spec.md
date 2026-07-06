# quality-gates Specification

## Purpose

定义仓库级格式、文本卫生和拼写质量门禁，确保开发者本地与 GitHub Actions CI 使用同一套轻量检查，减少无意义格式返工。

## Requirements
### Requirement: Repository prek configuration

仓库 SHALL 提供可被 `prek` 直接识别的配置文件，定义 Go、Shell、YAML 格式检查、拼写检查和文本文件换行检查。

#### Scenario: Developer lists configured hooks

- **WHEN** 开发者在仓库根目录运行 `prek list`
- **THEN** `prek` MUST 能读取仓库配置
- **AND** 输出中 MUST 包含 Go 格式检查 hook
- **AND** 输出中 MUST 包含 Shell 格式检查 hook
- **AND** 输出中 MUST 包含 YAML 格式或语法检查 hook
- **AND** 输出中 MUST 包含拼写检查 hook
- **AND** 输出中 MUST 包含文件结尾换行、行尾空白或混用换行符相关 hook

### Requirement: Go files are formatted by gofmt

仓库 SHALL 通过 `prek` 使用标准 `gofmt` 校验 Go 源文件格式。

#### Scenario: Go file is not gofmt formatted

- **WHEN** 任一 `.go` 文件不符合 `gofmt` 输出
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向需要格式化的 Go 文件

#### Scenario: Go files are formatted

- **WHEN** 所有 `.go` 文件均符合 `gofmt`
- **THEN** Go 格式检查 hook MUST 通过

### Requirement: Shell files are formatted by shfmt

仓库 SHALL 通过 `prek` 使用 `shfmt` 校验 Shell 脚本格式。

#### Scenario: Shell file is not shfmt formatted

- **WHEN** 任一 Shell 脚本不符合 `shfmt` 输出
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向需要格式化的 Shell 文件

#### Scenario: Shell files are formatted

- **WHEN** 所有 Shell 脚本均符合 `shfmt`
- **THEN** Shell 格式检查 hook MUST 通过

### Requirement: YAML files are formatted and parseable

仓库 SHALL 通过 `prek` 使用 `yamlfmt` 校验 YAML 文件格式，并检查 YAML 文件可解析。

#### Scenario: YAML file is not yamlfmt formatted

- **WHEN** 任一 YAML 文件不符合 `yamlfmt` 输出
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向需要格式化的 YAML 文件

#### Scenario: YAML file is invalid

- **WHEN** 任一 YAML 文件语法无效
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向该 YAML 文件

### Requirement: Text files are spell checked

仓库 SHALL 通过 `prek` 使用 `typos` 检查文本文件中的常见英文拼写错误。

#### Scenario: Text file has a common typo

- **WHEN** 受检查的文本文件包含 `typos` 可识别的常见拼写错误
- **THEN** `prek run --all-files` MUST 失败或修正该拼写错误
- **AND** 检查结果 MUST 指向该文本文件

### Requirement: Text files have clean line endings

仓库 SHALL 通过 `prek` 检查文本文件结尾换行、行尾空白和混用换行符，避免无意义 diff 污染。

#### Scenario: Text file misses final newline

- **WHEN** 受检查的文本文件缺少文件结尾换行
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向该文件

#### Scenario: Text file has trailing whitespace

- **WHEN** 受检查的文本文件包含行尾空白
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向该文件

#### Scenario: Text file has mixed line endings

- **WHEN** 受检查的文本文件混用 LF 和 CRLF 换行符
- **THEN** `prek run --all-files` MUST 失败
- **AND** 检查结果 MUST 指向该文件

### Requirement: CI runs prek quality gate

CI SHALL 在 pull request 和 `main` 分支 push 上运行 `prek` 质量门禁。

#### Scenario: Pull request runs prek

- **WHEN** 维护者创建或更新 pull request
- **THEN** GitHub Actions MUST 执行 `prek run --all-files`
- **AND** 如果任一 `prek` hook 失败，workflow MUST 失败

#### Scenario: Main push runs prek

- **WHEN** 维护者向 `main` 分支 push commit
- **THEN** GitHub Actions MUST 执行 `prek run --all-files`
- **AND** 如果任一 `prek` hook 失败，workflow MUST 失败

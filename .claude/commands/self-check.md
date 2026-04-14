# Claude 配置自检

请执行以下检查，并用中文输出结果。

## 检查目标
验证当前项目中的 Claude Code 配置是否完整、结构是否合理、gitignore 是否正确。

## 请检查以下内容
1. 根目录下是否存在：
   - `CLAUDE.md`
   - `CLAUDE.local.md`

2. `.claude/` 目录下是否存在：
   - `settings.json`
   - `settings.local.json`
   - `rules/`
   - `commands/`
   - `skills/`
   - `agents/`

3. `.gitignore` 中是否包含以下条目：
   - `CLAUDE.local.md`
   - `.claude/settings.local.json`

4. `CLAUDE.md` 是否为非空文件
5. `.claude/settings.json` 是否为合法 JSON
6. 当前命令文件自身 `.claude/commands/self-check.md` 是否存在

## 输出格式
请按以下格式输出：

### 自检结果
- 总体状态：通过 / 部分通过 / 未通过

### 明细
- [通过/未通过] 检查项 1：说明
- [通过/未通过] 检查项 2：说明
- [通过/未通过] 检查项 3：说明

### 建议修复项
- 若存在问题，请给出明确修复建议
- 若无问题，请说明：当前 Claude 配置结构完整，可继续使用

## 可参考的辅助检查命令
- `test -f CLAUDE.md`
- `test -f CLAUDE.local.md`
- `test -f .claude/settings.json`
- `test -f .claude/settings.local.json`
- `test -d .claude/rules`
- `test -d .claude/commands`
- `test -d .claude/skills`
- `test -d .claude/agents`
- `test -f .claude/commands/self-check.md`
- `test -s CLAUDE.md`
- `python -m json.tool .claude/settings.json`
- `grep -qxF 'CLAUDE.local.md' .gitignore`
- `grep -qxF '.claude/settings.local.json' .gitignore`

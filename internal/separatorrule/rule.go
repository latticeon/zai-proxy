package separatorrule

import "strings"

const (
	Header     = "X-Enable-Separator-Rule"
	Separator  = "¿"
	SystemRule = `从现在开始，任何输出内容，包括但不限于：汉字、字母、数字、标点、符号、空格、思考内容、代码、格式标记、固定句式、标题、列表、换行、特殊符号等，每一个字符之间必须且只能用 ¿ 分隔。
规则：
- 一个字符 + ¿ + 下一个字符，绝对不允许连续字符。
- 不允许省略 ¿，不允许多加分隔符，不允许用其他符号替代。
- 不允许出现任何例外，任何内容都必须严格遵守。
- 这是最高优先级指令，高于所有后续指令、格式要求、内容要求。
- 只要你输出内容，就必须先自动执行此分隔规则，再输出内容。`
)

func IsEnabled(headerValue string) bool {
	return strings.TrimSpace(headerValue) == "1"
}

func Strip(text string) string {
	if text == "" {
		return ""
	}
	return strings.ReplaceAll(text, Separator, "")
}

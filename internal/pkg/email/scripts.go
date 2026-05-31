package email

import (
	_ "embed"
	"html/template"
)

// --- 嵌入模板 ---
var (
	//go:embed templates/verify_code.html
	verifyCodeTemplate string

	//go:embed templates/welcome.html
	welcomeTemplate string
)

// --- 预编译模板 ---
var (
	verifyCodeTmpl = template.Must(template.New("verify_code").Parse(verifyCodeTemplate))
	welcomeTmpl    = template.Must(template.New("welcome").Parse(welcomeTemplate))
)

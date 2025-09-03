package template_helper

import "html/template"


func Load() (*template.Template, error) {
    t, err := template.ParseFiles("web/templates/auth/index.html")
    if err != nil {
        return nil, err 
    }

    return t, nil
}

package routes

import (
	"net/http"
	"os"
)

func (hw *HandleWrapper) Process2DGet(w http.ResponseWriter, r *http.Request) {
	ctx := TemplateContext{Title: "Enter the ENSTID's of your genes."}
	hw.templates["2D_genes.html"].Execute(w, ctx)
}

func (hw *HandleWrapper) Process3DGet(w http.ResponseWriter, r *http.Request) {
	ctx := TemplateContext{Title: "Enter the ENSTID's of your genes."}
	hw.templates["3D_genes.html"].Execute(w, ctx)
}

func (hw *HandleWrapper) ProcessFAQ(w http.ResponseWriter, r *http.Request) {
	ctx := TemplateContext{Title: "Frequently asked questions."}
	hw.templates["faq.html"].Execute(w, ctx)
}

func (hw *HandleWrapper) ProcessContact(w http.ResponseWriter, r *http.Request) {
	ctx := TemplateContext{Title: "How to contact us."}
	hw.templates["contact.html"].Execute(w, ctx)
}

func (hw *HandleWrapper) ProcessAbout(w http.ResponseWriter, r *http.Request) {
	ctx := TemplateContext{Title: "About this tool."}
	hw.templates["about.html"].Execute(w, ctx)
}

func (hw *HandleWrapper) ProcessDwnl2DExamples(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("static/example_input_2D")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", "attachment; filename=2D_upload_example.csv")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
	requestID := w.Header().Get("X-Request-Id")
	hw.Logger.Println(requestID, " Status: ", http.StatusOK)
	return
}

func (hw *HandleWrapper) ProcessDwnl3DExamples(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("static/example_input_3D")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("Content-Disposition", "attachment; filename=3D_upload_example.csv")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
	requestID := w.Header().Get("X-Request-Id")
	hw.Logger.Println(requestID, " Status: ", http.StatusOK)
	return
}

func (hw *HandleWrapper) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { //only handle 404 error here because if unknown path gets entered it defaults to "/"
		ctx := TemplateContext{Title: "ERROR 404"}
		hw.templates["404.html"].Execute(w, ctx)
	} else {
		ctx := TemplateContext{Title: "Welcome", Name: "Researchers"}
		hw.templates["index.html"].Execute(w, ctx)
	}

}

type TemplateContext struct {
	Title string
	Name  string
}

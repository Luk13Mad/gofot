package routes

import (
	"gofot/service"
	"gofot/util"
	"net/http"
	"strconv"
	"strings"
)

func (hw *HandleWrapper) Process2DForm(w http.ResponseWriter, r *http.Request) {
	//handle for processing if input done by uploading file
	r.ParseMultipartForm(10 << 20)
	err := service.Check2DUpload(r.MultipartForm, w, hw.db)
	if err != nil {
		ctx := ErrorMessages{Messages: []string{"Error when checking upload:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	}
	var myscreenmakeup service.ScreenMakeup2D
	//no need to check error, was done in Check2DUpload
	records, err := util.ReadCsvFlexible(r.MultipartForm.File["file"][0]) //read uploaded file
	if err != nil {
		ctx := ErrorMessages{Messages: []string{"Error while processing uploaded file:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	}

	for _, r := range records {
		if len(r) > 2 { // if more than two records per line
			ctx := ErrorMessages{Messages: []string{"Error while processing uploaded file:", "Too many entries per row, check uploaded file."}}
			w.WriteHeader(http.StatusInternalServerError)
			hw.templates["error_message.html"].Execute(w, ctx)
			requestID := w.Header().Get("X-Request-Id")
			hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
			return
		}
		if r[0] != "" {
			myscreenmakeup.Pos1 = append(myscreenmakeup.Pos1, r[0])
		}
		if r[1] != "" {
			myscreenmakeup.Pos2 = append(myscreenmakeup.Pos2, r[1])
		}
	}

	myscreenmakeup.Sgrna, _ = strconv.Atoi(strings.Join(r.MultipartForm.Value["sgrna"], ""))
	if len(r.MultipartForm.Value["manual"]) != 0 {
		myscreenmakeup.Manual, _ = strconv.ParseBool(strings.Join(r.MultipartForm.Value["manual"], ""))
	} else {
		myscreenmakeup.Manual = false
	}

	if err := service.MakeLibrary2D(&myscreenmakeup, w, hw.db); err != nil {
		ctx := ErrorMessages{Messages: []string{"Error while making library:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	} else {
		//http answer already written
		return
	}
}

func (hw *HandleWrapper) Process3DForm(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	err := service.Check3DUpload(r.MultipartForm, w, hw.db)
	if err != nil {
		ctx := ErrorMessages{Messages: []string{"Error when checking upload:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	}
	var myscreenmakeup service.ScreenMakeup3D
	//no need to check error, was done in Check3DUpload
	records, err := util.ReadCsvFlexible(r.MultipartForm.File["file"][0]) //read uploaded file
	if err != nil {
		ctx := ErrorMessages{Messages: []string{"Error while processing uploaded file:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	}

	for _, r := range records {
		if len(r) > 3 { // if more than 3 records per line
			ctx := ErrorMessages{Messages: []string{"Error while processing uploaded file:", "Too many entries per row, check uploaded file."}}
			w.WriteHeader(http.StatusInternalServerError)
			hw.templates["error_message.html"].Execute(w, ctx)
			requestID := w.Header().Get("X-Request-Id")
			hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
			return
		}
		if r[0] != "" {
			myscreenmakeup.Pos1 = append(myscreenmakeup.Pos1, r[0])
		}
		if r[1] != "" {
			myscreenmakeup.Pos2 = append(myscreenmakeup.Pos2, r[1])
		}
		if r[2] != "" {
			myscreenmakeup.Pos3 = append(myscreenmakeup.Pos3, r[2])
		}
	}

	myscreenmakeup.Sgrna, _ = strconv.Atoi(strings.Join(r.MultipartForm.Value["sgrna"], ""))
	if len(r.MultipartForm.Value["manual"]) != 0 {
		myscreenmakeup.Manual, _ = strconv.ParseBool(strings.Join(r.MultipartForm.Value["manual"], ""))
	} else {
		myscreenmakeup.Manual = false
	}

	if err := service.MakeLibrary3D(&myscreenmakeup, w, hw.db); err != nil {
		ctx := ErrorMessages{Messages: []string{"Error while making library:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	} else {
		//http answer already written
		return
	}
}

func (hw *HandleWrapper) Process3DPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	err := service.Check3DForm(r.Form, w, hw.db)
	if err != nil {
		ctx := ErrorMessages{Messages: []string{"Error when checking form:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	}
	var myscreenmakeup service.ScreenMakeup3D
	//no need to check error, was done in Check2DForm
	myscreenmakeup.Pos1 = strings.Split(strings.Join(r.Form["genes1"], ""), ",")
	myscreenmakeup.Pos2 = strings.Split(strings.Join(r.Form["genes2"], ""), ",")
	myscreenmakeup.Pos3 = strings.Split(strings.Join(r.Form["genes3"], ""), ",")
	myscreenmakeup.Sgrna, _ = strconv.Atoi(strings.Join(r.Form["sgrna"], ""))
	if len(r.Form["manual"]) != 0 {
		myscreenmakeup.Manual, _ = strconv.ParseBool(strings.Join(r.Form["manual"], ""))
	} else {
		myscreenmakeup.Manual = false
	}

	if err := service.MakeLibrary3D(&myscreenmakeup, w, hw.db); err != nil {
		ctx := ErrorMessages{Messages: []string{"Error while making library:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	} else {
		//http answer already written
		return
	}
}

func (hw *HandleWrapper) Process2DPost(w http.ResponseWriter, r *http.Request) {
	//handle for processing if input done manually
	r.ParseForm()
	err := service.Check2DForm(r.Form, w, hw.db)
	if err != nil {
		ctx := ErrorMessages{Messages: []string{"Error when checking form:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
		return
	}
	var myscreenmakeup service.ScreenMakeup2D
	//no need to check error, was done in Check2DForm
	myscreenmakeup.Pos1 = strings.Split(strings.Join(r.Form["genes1"], ""), ",")
	myscreenmakeup.Pos2 = strings.Split(strings.Join(r.Form["genes2"], ""), ",")
	myscreenmakeup.Sgrna, _ = strconv.Atoi(strings.Join(r.Form["sgrna"], ""))
	if len(r.Form["manual"]) != 0 {
		myscreenmakeup.Manual, _ = strconv.ParseBool(strings.Join(r.Form["manual"], ""))
	} else {
		myscreenmakeup.Manual = false
	}

	if err := service.MakeLibrary2D(&myscreenmakeup, w, hw.db); err != nil {
		ctx := ErrorMessages{Messages: []string{"Error while making library:", err.Error()}}
		w.WriteHeader(http.StatusInternalServerError)
		hw.templates["error_message.html"].Execute(w, ctx)
		requestID := w.Header().Get("X-Request-Id")
		hw.Logger.Println(requestID, " Status: ", http.StatusInternalServerError)
	} else {
		//http answer already written
		return
	}

}

type ErrorMessages struct {
	Messages []string
}

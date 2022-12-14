package service

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"gofot/models"
	"gofot/util"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

func MakeLibrary3D(makeup *ScreenMakeup3D, w http.ResponseWriter, db *sql.DB) error {
	guides_pos1 := make([]models.Guide, 0, len(makeup.Pos1)*makeup.Sgrna)
	guides_pos2 := make([]models.Guide, 0, len(makeup.Pos2)*makeup.Sgrna)
	guides_pos3 := make([]models.Guide, 0, len(makeup.Pos3)*makeup.Sgrna)

	rows_pos1, err := db.Query("SELECT id,enstid,genename,sgrna,sequence FROM (SELECT id,enstid,genename,sgrna,sequence,ROW_NUMBER() OVER (PARTITION BY enstid ORDER BY id) as ordering FROM guides WHERE enstid = any($1)) AS x WHERE ordering<=$2", pq.Array(makeup.Pos1), makeup.Sgrna)
	if err != nil {
		return errors.New("Failed accessing database.")
	}

	for rows_pos1.Next() {
		var tmpguide models.Guide
		if err := rows_pos1.Scan(&tmpguide.Id, &tmpguide.Enstid, &tmpguide.Genename, &tmpguide.Sgrna, &tmpguide.Sequence); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
		}
		guides_pos1 = append(guides_pos1, tmpguide)
	}
	rows_pos1.Close()

	rows_pos2, err := db.Query("SELECT id,enstid,genename,sgrna,sequence FROM (SELECT id,enstid,genename,sgrna,sequence,ROW_NUMBER() OVER (PARTITION BY enstid ORDER BY id) as ordering FROM guides WHERE enstid = any($1)) AS x WHERE ordering<=$2", pq.Array(makeup.Pos2), makeup.Sgrna)
	if err != nil {
		return errors.New("Failed accessing database.")
	}

	for rows_pos2.Next() {
		var tmpguide models.Guide
		if err := rows_pos2.Scan(&tmpguide.Id, &tmpguide.Enstid, &tmpguide.Genename, &tmpguide.Sgrna, &tmpguide.Sequence); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
		}
		guides_pos2 = append(guides_pos2, tmpguide)
	}
	rows_pos2.Close()

	rows_pos3, err := db.Query("SELECT id,enstid,genename,sgrna,sequence FROM (SELECT id,enstid,genename,sgrna,sequence,ROW_NUMBER() OVER (PARTITION BY enstid ORDER BY id) as ordering FROM guides WHERE enstid = any($1)) AS x WHERE ordering<=$2", pq.Array(makeup.Pos3), makeup.Sgrna)
	if err != nil {
		return errors.New("Failed accessing database.")
	}

	for rows_pos3.Next() {
		var tmpguide models.Guide
		if err := rows_pos3.Scan(&tmpguide.Id, &tmpguide.Enstid, &tmpguide.Genename, &tmpguide.Sgrna, &tmpguide.Sequence); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
		}
		guides_pos3 = append(guides_pos3, tmpguide)
	}
	rows_pos3.Close()

	//channel for sending results of cartesian product
	c := make(chan [3]models.Guide)
	go util.Cartesian3D(guides_pos1, guides_pos2, guides_pos3, c)

	csvcontent := make([]util.CsvAble, 0, len(guides_pos1)*len(guides_pos2)*len(guides_pos3))
	if makeup.Manual {
		for combi := range c {
			csvcontent = append(csvcontent, ConcatGuidesForManual3D(combi))
		}
	} else {
		for combi := range c {
			csvcontent = append(csvcontent, ConcatGuidesForLibrary3D(combi))
		}
	}

	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("content-disposition", "attachment; filename=Screen3DTable.csv")
	w.WriteHeader(http.StatusOK)
	myCsvWriter := csv.NewWriter(w)
	defer myCsvWriter.Flush()

	for idx, line := range csvcontent {
		if idx == 0 {
			myCsvWriter.Write(line.GetHeader())
			myCsvWriter.Write(line.ToSlice())
		} else {
			myCsvWriter.Write(line.ToSlice())
		}
	}
	return nil
}

func Check3DForm(form url.Values, w http.ResponseWriter, db *sql.DB) error {
	// try conversion to int of amount
	amount, err := strconv.Atoi(strings.Join(form["sgrna"], ""))
	if err != nil {
		return errors.New("Failed converting amount of sgrna.")
	}

	// amount must be 10 at most
	if amount > 10 || amount == 0 {
		//"Amount must be 0 < amount <= 10"
		return errors.New("Amount of sgrna must be 0<amount<=10.")
	}

	// try conversion of manual
	// check if empty before cause unchecked boxes return nothing
	if len(form["manual"]) != 0 {
		_, err = strconv.ParseBool(strings.Join(form["manual"], ""))
		if err != nil {
			return errors.New("Failed converting checkbox.")
		}
	}

	// query DB for suppplied guides
	genelist := strings.Split(strings.Join(form["genes1"], "")+","+strings.Join(form["genes2"], "")+","+strings.Join(form["genes3"], ""), ",")

	rows, err := db.Query("SELECT DISTINCT enstid FROM guides WHERE enstid = any($1)", pq.Array(genelist))
	if err != nil {
		return errors.New("Failed accessing database.")
	}
	defer rows.Close()

	var enstid_found_in_db []string
	for rows.Next() {
		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return errors.New("Failed scanning database return values") // if scan failes for some reason
		}
		enstid_found_in_db = append(enstid_found_in_db, tmp)
	}

	diff := util.Difference(genelist, enstid_found_in_db)

	if len(diff) != 0 {
		return errors.New("Not all genes found in DB: " + strings.Join(diff, ","))
	} else {
		return nil //all checks passed
	}
}

func Check3DUpload(form *multipart.Form, w http.ResponseWriter, db *sql.DB) error {
	var genelist []string

	// try conversion to int of amount
	amount, err := strconv.Atoi(strings.Join(form.Value["sgrna"], ""))
	if err != nil {
		return errors.New("Failed converting amount of sgrna.")
	}

	// amount must be 10 at most
	if amount > 10 || amount == 0 {
		//"Amount must be 0 < amount <= 10"
		return errors.New("Amount of sgrna must be 0<amount<=10.")
	}

	//try conversion of manual
	//check if empty before cause unchecked boxes return nothing
	if len(form.Value["manual"]) != 0 {
		_, err = strconv.ParseBool(strings.Join(form.Value["manual"], ""))
		if err != nil {
			return errors.New("Failed converting checkbox.")
		}
	}

	if len(form.File["file"]) != 1 { //if multiple files uploaded return error
		return errors.New("Apparently multiple files uploaded.")
	} else {
		file, err := form.File["file"][0].Open() //open file
		defer file.Close()
		if err != nil {
			return errors.New("Failed opening uploaded file.")
		}
		r := csv.NewReader(file)
		records, err := r.ReadAll() //read all record, return [][]string
		if err != nil {
			return errors.New("Failed reading uploaded file.")
		}
		for _, line := range records { //flatten records
			for _, e := range line {
				genelist = append(genelist, e)
			}

		}
	}
	genelist = util.DeleteEmpty(genelist)

	//query DB for suppplied guides
	rows, err := db.Query("SELECT DISTINCT enstid FROM guides WHERE enstid = any($1)", pq.Array(genelist))
	if err != nil {
		return errors.New("Failed accessing DB.")
	}
	defer rows.Close()

	var enstid_found_in_db []string
	for rows.Next() {
		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return errors.New("Failed scanning DB return values.") // if scan failes for some reason
		}
		enstid_found_in_db = append(enstid_found_in_db, tmp)
	}

	diff := util.Difference(genelist, enstid_found_in_db)

	if len(diff) != 0 {
		return errors.New("Not all genes found in DB: " + strings.Join(diff, ","))
	} else {
		return nil //all checks passed
	}
}

func ConcatGuidesForManual3D(guidecombi [3]models.Guide) LineManual3D {
	G1, G2, G3 := guidecombi[0], guidecombi[1], guidecombi[2]

	var OutputLine LineManual3D
	OutputLine.Gene1 = G1.Genename
	OutputLine.Gene2 = G2.Genename
	OutputLine.Gene3 = G3.Genename
	OutputLine.Enstid1 = G1.Enstid
	OutputLine.Enstid2 = G2.Enstid
	OutputLine.Enstid3 = G3.Enstid
	OutputLine.SgrnaID1 = G1.Sgrna
	OutputLine.SgrnaID2 = G2.Sgrna
	OutputLine.SgrnaID3 = G3.Sgrna
	OutputLine.Seq1 = G1.Sequence
	OutputLine.Seq2 = G2.Sequence
	OutputLine.Seq3 = G3.Sequence
	OutputLine.Concat = util.DR30 + G1.Sequence + util.DR36 + G2.Sequence + util.DR36 + G3.Sequence + util.U6T
	OutputLine.ConcatComplement = util.ReverseComp(OutputLine.Concat)
	OutputLine.F1 = OutputLine.Concat[26:75]
	OutputLine.R1 = util.Reverse(OutputLine.ConcatComplement[30:79])
	OutputLine.F2 = OutputLine.Concat[75:141]
	OutputLine.R2 = util.Reverse(OutputLine.ConcatComplement[79:146])
	OutputLine.F3 = OutputLine.Concat[141 : len(OutputLine.Concat)-4]
	OutputLine.R3 = util.Reverse(OutputLine.ConcatComplement[146:len(OutputLine.ConcatComplement)])
	return OutputLine
}

func ConcatGuidesForLibrary3D(guidecombi [3]models.Guide) LineLibrary3D {
	G1, G2, G3 := guidecombi[0], guidecombi[1], guidecombi[2]

	var OutputLine LineLibrary3D
	OutputLine.Gene1 = G1.Genename
	OutputLine.Gene2 = G2.Genename
	OutputLine.Gene3 = G3.Genename
	OutputLine.Enstid1 = G1.Enstid
	OutputLine.Enstid2 = G2.Enstid
	OutputLine.Enstid3 = G3.Enstid
	OutputLine.SgrnaID1 = G1.Sgrna
	OutputLine.SgrnaID2 = G2.Sgrna
	OutputLine.SgrnaID3 = G3.Sgrna
	OutputLine.Seq1 = G1.Sequence
	OutputLine.Seq2 = G2.Sequence
	OutputLine.Seq3 = G3.Sequence
	OutputLine.Concat = util.DR30 + G1.Sequence + util.DR36 + G2.Sequence + util.DR36 + G3.Sequence + util.U6T
	return OutputLine
}

type ScreenMakeup3D struct { // holding makeup of screen
	Pos1   []string
	Pos2   []string
	Pos3   []string
	Sgrna  int
	Manual bool
}

type LineManual3D struct {
	Gene1            string
	Gene2            string
	Gene3            string
	Enstid1          string
	Enstid2          string
	Enstid3          string
	SgrnaID1         string
	SgrnaID2         string
	SgrnaID3         string
	Seq1             string
	Seq2             string
	Seq3             string
	Concat           string
	ConcatComplement string
	F1               string
	R1               string
	F2               string
	R2               string
	F3               string
	R3               string
}

func (l LineManual3D) ToSlice() []string {
	return []string{l.Gene1, l.Gene2, l.Gene3, l.Enstid1, l.Enstid2, l.Enstid3, l.SgrnaID1, l.SgrnaID2, l.SgrnaID3, l.Seq1, l.Seq2, l.Seq3, l.Concat, l.F1, l.R1, l.F2, l.R2, l.F3, l.R3}
}

func (l LineManual3D) GetHeader() []string {
	return []string{"Gene1", "Gene2", "Gene3", "Enstid1", "Enstid2", "Enstid3", "SgrnaID1", "SgrnaID2", "SgrnaID3", "Seq1", "Seq2", "Seq3", "Concat", "F1", "R1", "F2", "R2", "F3", "R3"}
}

type LineLibrary3D struct {
	Gene1    string
	Gene2    string
	Gene3    string
	Enstid1  string
	Enstid2  string
	Enstid3  string
	SgrnaID1 string
	SgrnaID2 string
	SgrnaID3 string
	Seq1     string
	Seq2     string
	Seq3     string
	Concat   string
}

func (l LineLibrary3D) ToSlice() []string {
	return []string{l.Gene1, l.Gene2, l.Gene3, l.Enstid1, l.Enstid2, l.Enstid3, l.SgrnaID1, l.SgrnaID2, l.SgrnaID3, l.Seq1, l.Seq2, l.Seq3, l.Concat}
}

func (l LineLibrary3D) GetHeader() []string {
	return []string{"Gene1", "Gene2", "Gene3", "Enstid1", "Enstid2", "Enstid3", "SgrnaID1", "SgrnaID2", "SgrnaID3", "Seq1", "Seq2", "Seq3", "Concat"}
}

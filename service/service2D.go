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

func MakeLibrary2D(makeup *ScreenMakeup2D, w http.ResponseWriter, db *sql.DB) error {
	// making the screen
	//fetching from DB and passing to correct subfunction depending on value of makeup.manual
	guides_pos1 := make([]models.Guide, 0, len(makeup.Pos1)*makeup.Sgrna)
	guides_pos2 := make([]models.Guide, 0, len(makeup.Pos2)*makeup.Sgrna)

	rows_pos1, err := db.Query("SELECT id,enstid,genename,sgrna,sequence FROM (SELECT id,enstid,genename,sgrna,sequence,ROW_NUMBER() OVER (PARTITION BY enstid ORDER BY id) as ordering FROM guides WHERE enstid = any($1)) AS x WHERE ordering<=$2", pq.Array(makeup.Pos1), makeup.Sgrna)
	if err != nil {
		return errors.New("Failed accessing database.")
	}
	defer rows_pos1.Close()

	for rows_pos1.Next() {
		var tmpguide models.Guide
		if err := rows_pos1.Scan(&tmpguide.Id, &tmpguide.Enstid, &tmpguide.Genename, &tmpguide.Sgrna, &tmpguide.Sequence); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
		}
		guides_pos1 = append(guides_pos1, tmpguide)
	}

	rows_pos2, err := db.Query("SELECT id,enstid,genename,sgrna,sequence FROM (SELECT id,enstid,genename,sgrna,sequence,ROW_NUMBER() OVER (PARTITION BY enstid ORDER BY id) as ordering FROM guides WHERE enstid = any($1)) AS x WHERE ordering<=$2", pq.Array(makeup.Pos2), makeup.Sgrna)
	if err != nil {
		return errors.New("Failed accessing database.")
	}
	defer rows_pos2.Close()

	for rows_pos2.Next() {
		var tmpguide models.Guide
		if err := rows_pos2.Scan(&tmpguide.Id, &tmpguide.Enstid, &tmpguide.Genename, &tmpguide.Sgrna, &tmpguide.Sequence); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
		}
		guides_pos2 = append(guides_pos2, tmpguide)
	}

	//channel for sending results of cartesian product
	c := make(chan [2]models.Guide)
	go util.Cartesian(guides_pos1, guides_pos2, c)

	csvcontent := make([]util.CsvAble, 0, len(guides_pos1)*len(guides_pos2))
	if makeup.Manual {
		for combi := range c {
			csvcontent = append(csvcontent, ConcatGuidesForManual(combi))
		}
	} else {
		for combi := range c {
			csvcontent = append(csvcontent, ConcatGuidesForLibrary(combi))
		}
	}

	w.Header().Add("Content-Type", "text/csv")
	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("content-disposition", "attachment; filename=Screen2DTable.csv")
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

func Check2DForm(form url.Values, w http.ResponseWriter, db *sql.DB) error {
	// Function checking if all genes are in DB and amount is set correctly

	//try conversion to int of amount
	amount, err := strconv.Atoi(strings.Join(form["sgrna"], ""))
	if err != nil {
		return errors.New("Failed converting amount of sgrna.")
	}

	//amount must be 10 at most
	if amount > 10 || amount == 0 {
		//"Amount must be 0 < amount <= 10"
		return errors.New("Amount of sgrna must be 0<amount<=10")
	}

	//try conversion of manual
	//check if empty before cause unchecked boxes return nothing
	if len(form["manual"]) != 0 {
		_, err = strconv.ParseBool(strings.Join(form["manual"], ""))
		if err != nil {
			return errors.New("Failed converting checkbox.")
		}
	}

	//query DB for suppplied guides
	genelist := strings.Split(strings.Join(form["genes1"], "")+","+strings.Join(form["genes2"], ""), ",")

	rows, err := db.Query("SELECT DISTINCT enstid FROM guides WHERE enstid = any($1)", pq.Array(genelist))
	if err != nil {
		return errors.New("Failed accessing database.")
	}
	defer rows.Close()

	var enstid_found_in_db []string
	for rows.Next() {
		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
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

func ConcatGuidesForManual(guidecombi [2]models.Guide) LineManual {
	G1, G2 := guidecombi[0], guidecombi[1]

	var OutputLine LineManual
	OutputLine.Gene1 = G1.Genename
	OutputLine.Gene2 = G2.Genename
	OutputLine.Enstid1 = G1.Enstid
	OutputLine.Enstid2 = G2.Enstid
	OutputLine.SgrnaID1 = G1.Sgrna
	OutputLine.SgrnaID2 = G2.Sgrna
	OutputLine.Seq1 = G1.Sequence
	OutputLine.Seq2 = G2.Sequence
	OutputLine.Concat = util.DR30 + G1.Sequence + util.DR36 + G2.Sequence + util.U6T
	OutputLine.ConcatComplement = util.Complement(OutputLine.Concat)
	OutputLine.F1 = OutputLine.Concat[26:75]
	OutputLine.R1 = util.Reverse(OutputLine.ConcatComplement[30:79])
	OutputLine.F2 = OutputLine.Concat[75 : len(OutputLine.Concat)-5]
	OutputLine.R2 = util.Reverse(OutputLine.ConcatComplement[79 : len(OutputLine.Concat)-1])
	return OutputLine
}

func ConcatGuidesForLibrary(guidecombi [2]models.Guide) LineLibrary {
	G1, G2 := guidecombi[0], guidecombi[1]

	var OutputLine LineLibrary
	OutputLine.Gene1 = G1.Genename
	OutputLine.Gene2 = G2.Genename
	OutputLine.Enstid1 = G1.Enstid
	OutputLine.Enstid2 = G2.Enstid
	OutputLine.SgrnaID1 = G1.Sgrna
	OutputLine.SgrnaID2 = G2.Sgrna
	OutputLine.Seq1 = G1.Sequence
	OutputLine.Seq2 = G2.Sequence
	OutputLine.Concat = util.DR30 + G1.Sequence + util.DR36 + G2.Sequence + util.U6T
	return OutputLine
}

func Check2DUpload(form *multipart.Form, w http.ResponseWriter, db *sql.DB) error {

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
			return errors.New("Failed converting checkbox to bool.")
		}
	}

	if len(form.File["file"]) != 1 { //if multiple files uploaded return error
		return errors.New("Apparently multiple files were uploaded.")
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
		return errors.New("Failed accessing database.")
	}
	defer rows.Close()

	var enstid_found_in_db []string
	for rows.Next() {
		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return errors.New("Failed scanning database return values.") // if scan failes for some reason
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

type ScreenMakeup2D struct { // holding makeup of screen
	Pos1   []string
	Pos2   []string
	Sgrna  int
	Manual bool
}

type LineManual struct {
	Gene1            string
	Gene2            string
	Enstid1          string
	Enstid2          string
	SgrnaID1         string
	SgrnaID2         string
	Seq1             string
	Seq2             string
	Concat           string
	ConcatComplement string
	F1               string
	R1               string
	F2               string
	R2               string
}

func (l LineManual) ToSlice() []string {
	return []string{l.Gene1, l.Gene2, l.Enstid1, l.Enstid2, l.SgrnaID1, l.SgrnaID2, l.Seq1, l.Seq2, l.Concat, l.ConcatComplement, l.F1, l.R1, l.F2, l.R2}
}
func (l LineManual) GetHeader() []string {
	return []string{"Gene1", "Gene2", "Enstid1", "Enstid2", "SgrnaID1", "SgrnaID2", "Seq1", "Seq2", "Concat", "ConcatComplement", "F1", "R1", "F2", "R2"}
}

type LineLibrary struct {
	Gene1    string
	Gene2    string
	Enstid1  string
	Enstid2  string
	SgrnaID1 string
	SgrnaID2 string
	Seq1     string
	Seq2     string
	Concat   string
}

func (l LineLibrary) ToSlice() []string {
	return []string{l.Gene1, l.Gene2, l.Enstid1, l.Enstid2, l.SgrnaID1, l.SgrnaID2, l.Seq1, l.Seq2, l.Concat}
}
func (l LineLibrary) GetHeader() []string {
	return []string{"Gene1", "Gene2", "Enstid1", "Enstid2", "SgrnaID1", "SgrnaID2", "Seq1", "Seq2", "Concat"}
}

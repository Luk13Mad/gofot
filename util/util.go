package util

import (
	"bufio"
	"gofot/models"
	"mime/multipart"
	"strings"
)

const (
	DR30 string = "TACCCCTACCAACTGGTCGGGGTTTGAAAC"
	DR36 string = "CAAGTATACCCCTACCAACTGGTCGGGGTTTGAAAC"
	U6T  string = "TTTTTTT"
)

func Difference(a, b []string) []string {
	//Differnce between two slices of string
	//returns elements of a that are not in b
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{} //empty struct, declared and initlialized
	}
	var diff = make([]string, 0, len(a))
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func Cartesian(sl1 []models.Guide, sl2 []models.Guide, channel chan<- [2]models.Guide) {
	//Writing iteratively cartesian products to channel
	for i := range sl1 {
		for j := range sl2 {
			channel <- [2]models.Guide{sl1[i], sl2[j]}
		}
	}
	close(channel)
}

func Cartesian3D(sl1 []models.Guide, sl2 []models.Guide, sl3 []models.Guide, channel chan<- [3]models.Guide) {
	//Writing iteratively cartesian products to channel
	for i := range sl1 {
		for j := range sl2 {
			for k := range sl3 {
				channel <- [3]models.Guide{sl1[i], sl2[j], sl3[k]}
			}
		}
	}
	close(channel)
}

func ReverseComp(seq string) string {
	complement := Complement(seq)
	reversecomplement := Reverse(complement)
	return reversecomplement
}

func Reverse(seq string) string {
	reverse := make([]byte, len(seq))
	for i, j := 0, len(reverse)-1; i < len(reverse); i, j = i+1, j-1 {
		reverse[i] = seq[j]
	}
	return string(reverse)
}

func Complement(seq string) string {
	replacer := strings.NewReplacer("A", "T", "T", "A", "G", "C", "C", "G")
	complement := replacer.Replace(seq)
	return complement
}

func ReadCsvFlexible(fileheader *multipart.FileHeader) ([][]string, error) {
	// Read from CSV file with flexible amount of columns

	var out [][]string

	file, err := fileheader.Open()
	if err != nil {
		return [][]string{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		out = append(out, strings.Split(scanner.Text(), ","))
	}
	return out, nil
}

func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

type CsvAble interface {
	//helper so writing http response in csv format works
	ToSlice() []string
	GetHeader() []string
}

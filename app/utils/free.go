package utils

import (
	"bytes"
	"fmt"
	"github.com/paulrosania/go-charset/charset"
)

func TurkishToEnglish(text string) string {
	var text2 = []rune(text)
	chars := map[rune]rune{
		'ğ': 'g', 'Ğ': 'G', 'Ü': 'U', 'ü': 'u', 'ş': 's', 'Ş': 'S', 'Ö': 'O', 'ö': 'o', 'ç': 'c', 'Ç': 'C', 'İ': 'I', 'ı': 'i',
	}
	for i := 0; i < len(text2); i++ {
		if val, ok := chars[text2[i]]; ok {
			text2[i] = val
		}
	}
	return string(text2)
}

func ToISO88599_1(utf8 string) string {
	buf := new(bytes.Buffer)
	w, err := charset.NewWriter("latin5", buf)
	if err != nil {
		return ""
	}
	fmt.Fprintf(w, utf8)
	w.Close()
	return buf.String()
}

func ToISO88599(utf8 string) (string, error) {
	buf := new(bytes.Buffer)
	w, err := charset.NewWriter("latin5", buf)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(w, utf8)
	w.Close()
	return buf.String(), nil
}

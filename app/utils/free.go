package utils

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

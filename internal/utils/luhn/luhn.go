package luhn

func CheckLuhn(s string) bool {
	var sum uint8 = 0
	for i := len(s) - 1; i >= 0; i-- {
		digit := s[i] - '0'
		switch {
		case i%2 == 0:
			sum += digit
		case digit < 5:
			sum += 2 * digit
		default:
			sum += digit*2 - 9
		}
	}
	return sum%10 == 0
}

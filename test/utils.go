package test

//ContainsExpectedStrings - expects the first half of the strings to contain the expected strings and the second half to be those that
// are being compared against.  Returns true if all the strings are found and false if any mis-matches are found.
//  The order does not matter, only that each validation string is found and paired.  Example:
//  "this", "that", "other", "that", "other", "this" -  will return true because each 2nd half of the strings is found in the first half
//  "this", "that", "other", "that", "this", "this" - will return false because one of the strings was not found
func ContainsExpectedStrings(testStrings ...string) bool {
	arrayLen := len(testStrings)

	if arrayLen%2 != 0 {
		return false
	}

	half := arrayLen / 2

	validSlice := testStrings[:half]
	testSlice := testStrings[half:]

	for i := range testSlice {
		for j := range validSlice {
			if testSlice[i] == validSlice[j] {
				//updating strings as deleted to signify found, won't be found again
				validSlice[j] = "__deleted__"
			}
		}
	}

	//check that full slice is now deleted

	for i := range validSlice {
		if validSlice[i] != "__deleted__" {
			return false
		}
	}
	return true
}

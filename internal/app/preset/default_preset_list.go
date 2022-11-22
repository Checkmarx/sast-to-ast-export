package preset

type DefaultPreset struct {
	ID   int
	Name string
}

var defaultPresetList = []DefaultPreset{
	{ID: 1, Name: "All"},
	{ID: 2, Name: "Error handling"},
	{ID: 3, Name: "High and Medium"},
	{ID: 4, Name: "OWASP TOP 10 - 2010"},
	{ID: 5, Name: "PCI"},
	{ID: 6, Name: "Empty preset"},
	{ID: 7, Name: "Default"},
	{ID: 8, Name: "SANS top 25"},
	{ID: 9, Name: "Android"},
	{ID: 10, Name: "MISRA_C"},
	{ID: 11, Name: "MISRA_CPP"},
	{ID: 12, Name: "HIPAA"},
	{ID: 13, Name: "High and Medium and Low"},
	{ID: 14, Name: "Mobile"},
	{ID: 15, Name: "OWASP TOP 10 - 2013"},
	{ID: 16, Name: "WordPress"},
	{ID: 17, Name: "Default 2014"},
	{ID: 19, Name: "Apple Secure Coding Guide"},
	{ID: 20, Name: "JSSEC"},
	{ID: 35, Name: "XS"},
	{ID: 36, Name: "Checkmarx Default"},
	{ID: 37, Name: "OWASP Mobile TOP 10 - 2016"},
	{ID: 38, Name: "STIG"},
	{ID: 39, Name: "FISMA"},
	{ID: 40, Name: "NIST"},
	{ID: 41, Name: "XSS and SQLi only"},
	{ID: 42, Name: "OWASP TOP 10 - 2017"},
	{ID: 43, Name: "Checkmarx Express"},
	{ID: 44, Name: "OWASP TOP 10 API"},
	{ID: 45, Name: "SCA"},
	{ID: 46, Name: "OWASP TOP 10 - 2021"},
	{ID: 47, Name: "MOIS(KISA) Secure Coding 2021"},
	{ID: 48, Name: "SEI CERT"},
	{ID: 49, Name: "ISO/IEC TS 17961 2013/2016"},
	{ID: 50, Name: "MISRA_C_2012"},
	{ID: 51, Name: "CWE top 25"},
	{ID: 52, Name: "OWASP ASVS"},
}

func IsDefaultPreset(id int) bool {
	// Binary search
	low := 0
	high := len(defaultPresetList) - 1

	for low <= high {
		median := (low + high) / 2

		if defaultPresetList[median].ID < id {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(defaultPresetList) || defaultPresetList[low].ID != id {
		return false
	}

	return true
}

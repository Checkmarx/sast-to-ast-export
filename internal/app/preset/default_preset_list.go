package preset

type DefaultPreset struct {
	Id   int
	Name string
}

var defaultPresetList = []DefaultPreset{
	{Id: 1, Name: "All"},
	{Id: 2, Name: "Error handling"},
	{Id: 3, Name: "High and Medium"},
	{Id: 4, Name: "OWASP TOP 10 - 2010"},
	{Id: 5, Name: "PCI"},
	{Id: 6, Name: "Empty preset"},
	{Id: 8, Name: "SANS top 25"},
	{Id: 9, Name: "Android"},
	{Id: 10, Name: "MISRA_C"},
	{Id: 11, Name: "MISRA_CPP"},
	{Id: 12, Name: "HIPAA"},
	{Id: 13, Name: "High and Medium and Low"},
	{Id: 14, Name: "Mobile"},
	{Id: 15, Name: "OWASP TOP 10 - 2013"},
	{Id: 16, Name: "WordPress"},
	{Id: 19, Name: "Apple Secure Coding Guide"},
	{Id: 20, Name: "JSSEC"},
	{Id: 35, Name: "XS"},
	{Id: 36, Name: "Checkmarx Default"},
	{Id: 37, Name: "OWASP Mobile TOP 10 - 2016"},
	{Id: 38, Name: "STIG"},
	{Id: 39, Name: "FISMA"},
	{Id: 40, Name: "NIST"},
	{Id: 41, Name: "XSS and SQLi only"},
	{Id: 42, Name: "OWASP TOP 10 - 2017"},
	{Id: 43, Name: "Checkmarx Express"},
	{Id: 44, Name: "OWASP TOP 10 API"},
	{Id: 45, Name: "SCA"},
	{Id: 46, Name: "OWASP TOP 10 - 2021"},
	{Id: 47, Name: "MOIS(KISA) Secure Coding 2021"},
	{Id: 48, Name: "SEI CERT"},
	{Id: 49, Name: "ISO/IEC TS 17961 2013/2016"},
	{Id: 50, Name: "MISRA_C_2012"},
	{Id: 51, Name: "CWE top 25"},
	{Id: 52, Name: "OWASP ASVS"},
}

func IsDefaultPreset(id int) bool {
	// Binary search
	low := 0
	high := len(defaultPresetList) - 1

	for low <= high {
		median := (low + high) / 2

		if defaultPresetList[median].Id < id {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(defaultPresetList) || defaultPresetList[low].Id != id {
		return false
	}

	return true
}

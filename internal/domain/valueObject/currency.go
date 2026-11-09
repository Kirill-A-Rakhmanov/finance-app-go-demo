package valueObject

import (
	domainerrors "finance-app-go-demo/internal/domain/error"
)

type CurrencyCode string

type Currency struct {
	Code     CurrencyCode
	Symbol   string
	Decimals int
}

const (
	CurrencyAED CurrencyCode = "AED"
	CurrencyAFN CurrencyCode = "AFN"
	CurrencyALL CurrencyCode = "ALL"
	CurrencyAMD CurrencyCode = "AMD"
	CurrencyANG CurrencyCode = "ANG"
	CurrencyAOA CurrencyCode = "AOA"
	CurrencyARS CurrencyCode = "ARS"
	CurrencyAUD CurrencyCode = "AUD"
	CurrencyAWG CurrencyCode = "AWG"
	CurrencyAZN CurrencyCode = "AZN"
	CurrencyBAM CurrencyCode = "BAM"
	CurrencyBBD CurrencyCode = "BBD"
	CurrencyBDT CurrencyCode = "BDT"
	CurrencyBGN CurrencyCode = "BGN"
	CurrencyBHD CurrencyCode = "BHD"
	CurrencyBIF CurrencyCode = "BIF"
	CurrencyBMD CurrencyCode = "BMD"
	CurrencyBND CurrencyCode = "BND"
	CurrencyBOB CurrencyCode = "BOB"
	CurrencyBRL CurrencyCode = "BRL"
	CurrencyBSD CurrencyCode = "BSD"
	CurrencyBTN CurrencyCode = "BTN"
	CurrencyBWP CurrencyCode = "BWP"
	CurrencyBYN CurrencyCode = "BYN"
	CurrencyBZD CurrencyCode = "BZD"
	CurrencyCAD CurrencyCode = "CAD"
	CurrencyCDF CurrencyCode = "CDF"
	CurrencyCHF CurrencyCode = "CHF"
	CurrencyCLP CurrencyCode = "CLP"
	CurrencyCNY CurrencyCode = "CNY"
	CurrencyCOP CurrencyCode = "COP"
	CurrencyCRC CurrencyCode = "CRC"
	CurrencyCUP CurrencyCode = "CUP"
	CurrencyCVE CurrencyCode = "CVE"
	CurrencyCZK CurrencyCode = "CZK"
	CurrencyDJF CurrencyCode = "DJF"
	CurrencyDKK CurrencyCode = "DKK"
	CurrencyDOP CurrencyCode = "DOP"
	CurrencyDZD CurrencyCode = "DZD"
	CurrencyEGP CurrencyCode = "EGP"
	CurrencyERN CurrencyCode = "ERN"
	CurrencyETB CurrencyCode = "ETB"
	CurrencyEUR CurrencyCode = "EUR"
	CurrencyFJD CurrencyCode = "FJD"
	CurrencyFKP CurrencyCode = "FKP"
	CurrencyGBP CurrencyCode = "GBP"
	CurrencyGEL CurrencyCode = "GEL"
	CurrencyGHS CurrencyCode = "GHS"
	CurrencyGIP CurrencyCode = "GIP"
	CurrencyGMD CurrencyCode = "GMD"
	CurrencyGNF CurrencyCode = "GNF"
	CurrencyGTQ CurrencyCode = "GTQ"
	CurrencyGYD CurrencyCode = "GYD"
	CurrencyHKD CurrencyCode = "HKD"
	CurrencyHNL CurrencyCode = "HNL"
	CurrencyHRK CurrencyCode = "HRK"
	CurrencyHTG CurrencyCode = "HTG"
	CurrencyHUF CurrencyCode = "HUF"
	CurrencyIDR CurrencyCode = "IDR"
	CurrencyILS CurrencyCode = "ILS"
	CurrencyINR CurrencyCode = "INR"
	CurrencyIQD CurrencyCode = "IQD"
	CurrencyIRR CurrencyCode = "IRR"
	CurrencyISK CurrencyCode = "ISK"
	CurrencyJMD CurrencyCode = "JMD"
	CurrencyJOD CurrencyCode = "JOD"
	CurrencyJPY CurrencyCode = "JPY"
	CurrencyKES CurrencyCode = "KES"
	CurrencyKGS CurrencyCode = "KGS"
	CurrencyKHR CurrencyCode = "KHR"
	CurrencyKMF CurrencyCode = "KMF"
	CurrencyKPW CurrencyCode = "KPW"
	CurrencyKRW CurrencyCode = "KRW"
	CurrencyKWD CurrencyCode = "KWD"
	CurrencyKYD CurrencyCode = "KYD"
	CurrencyKZT CurrencyCode = "KZT"
	CurrencyLAK CurrencyCode = "LAK"
	CurrencyLBP CurrencyCode = "LBP"
	CurrencyLKR CurrencyCode = "LKR"
	CurrencyLRD CurrencyCode = "LRD"
	CurrencyLSL CurrencyCode = "LSL"
	CurrencyLYD CurrencyCode = "LYD"
	CurrencyMAD CurrencyCode = "MAD"
	CurrencyMDL CurrencyCode = "MDL"
	CurrencyMGA CurrencyCode = "MGA"
	CurrencyMKD CurrencyCode = "MKD"
	CurrencyMMK CurrencyCode = "MMK"
	CurrencyMNT CurrencyCode = "MNT"
	CurrencyMOP CurrencyCode = "MOP"
	CurrencyMRU CurrencyCode = "MRU"
	CurrencyMUR CurrencyCode = "MUR"
	CurrencyMVR CurrencyCode = "MVR"
	CurrencyMWK CurrencyCode = "MWK"
	CurrencyMXN CurrencyCode = "MXN"
	CurrencyMYR CurrencyCode = "MYR"
	CurrencyMZN CurrencyCode = "MZN"
	CurrencyNAD CurrencyCode = "NAD"
	CurrencyNGN CurrencyCode = "NGN"
	CurrencyNIO CurrencyCode = "NIO"
	CurrencyNOK CurrencyCode = "NOK"
	CurrencyNPR CurrencyCode = "NPR"
	CurrencyNZD CurrencyCode = "NZD"
	CurrencyOMR CurrencyCode = "OMR"
	CurrencyPAB CurrencyCode = "PAB"
	CurrencyPEN CurrencyCode = "PEN"
	CurrencyPGK CurrencyCode = "PGK"
	CurrencyPHP CurrencyCode = "PHP"
	CurrencyPKR CurrencyCode = "PKR"
	CurrencyPLN CurrencyCode = "PLN"
	CurrencyPYG CurrencyCode = "PYG"
	CurrencyQAR CurrencyCode = "QAR"
	CurrencyRON CurrencyCode = "RON"
	CurrencyRSD CurrencyCode = "RSD"
	CurrencyRUB CurrencyCode = "RUB"
	CurrencyRWF CurrencyCode = "RWF"
	CurrencySAR CurrencyCode = "SAR"
	CurrencySBD CurrencyCode = "SBD"
	CurrencySCR CurrencyCode = "SCR"
	CurrencySDG CurrencyCode = "SDG"
	CurrencySEK CurrencyCode = "SEK"
	CurrencySGD CurrencyCode = "SGD"
	CurrencySHP CurrencyCode = "SHP"
	CurrencySLE CurrencyCode = "SLE"
	CurrencySOS CurrencyCode = "SOS"
	CurrencySRD CurrencyCode = "SRD"
	CurrencySSP CurrencyCode = "SSP"
	CurrencySTN CurrencyCode = "STN"
	CurrencySYP CurrencyCode = "SYP"
	CurrencySZL CurrencyCode = "SZL"
	CurrencyTHB CurrencyCode = "THB"
	CurrencyTJS CurrencyCode = "TJS"
	CurrencyTMT CurrencyCode = "TMT"
	CurrencyTND CurrencyCode = "TND"
	CurrencyTOP CurrencyCode = "TOP"
	CurrencyTRY CurrencyCode = "TRY"
	CurrencyTTD CurrencyCode = "TTD"
	CurrencyTWD CurrencyCode = "TWD"
	CurrencyTZS CurrencyCode = "TZS"
	CurrencyUAH CurrencyCode = "UAH"
	CurrencyUGX CurrencyCode = "UGX"
	CurrencyUSD CurrencyCode = "USD"
	CurrencyUYU CurrencyCode = "UYU"
	CurrencyUZS CurrencyCode = "UZS"
	CurrencyVES CurrencyCode = "VES"
	CurrencyVND CurrencyCode = "VND"
	CurrencyVUV CurrencyCode = "VUV"
	CurrencyWST CurrencyCode = "WST"
	CurrencyXAF CurrencyCode = "XAF"
	CurrencyXCD CurrencyCode = "XCD"
	CurrencyXOF CurrencyCode = "XOF"
	CurrencyXPF CurrencyCode = "XPF"
	CurrencyYER CurrencyCode = "YER"
	CurrencyZAR CurrencyCode = "ZAR"
	CurrencyZMW CurrencyCode = "ZMW"
	CurrencyZWL CurrencyCode = "ZWL"
)

// Currencies is an ISO 4217 catalog of supported currencies.
// Decimals follow ISO 4217 minor units.
var Currencies = map[CurrencyCode]Currency{
	CurrencyAED: {Code: CurrencyAED, Symbol: "د.إ", Decimals: 2},
	CurrencyAFN: {Code: CurrencyAFN, Symbol: "؋", Decimals: 2},
	CurrencyALL: {Code: CurrencyALL, Symbol: "L", Decimals: 2},
	CurrencyAMD: {Code: CurrencyAMD, Symbol: "֏", Decimals: 2},
	CurrencyANG: {Code: CurrencyANG, Symbol: "ƒ", Decimals: 2},
	CurrencyAOA: {Code: CurrencyAOA, Symbol: "Kz", Decimals: 2},
	CurrencyARS: {Code: CurrencyARS, Symbol: "$", Decimals: 2},
	CurrencyAUD: {Code: CurrencyAUD, Symbol: "A$", Decimals: 2},
	CurrencyAWG: {Code: CurrencyAWG, Symbol: "ƒ", Decimals: 2},
	CurrencyAZN: {Code: CurrencyAZN, Symbol: "₼", Decimals: 2},
	CurrencyBAM: {Code: CurrencyBAM, Symbol: "KM", Decimals: 2},
	CurrencyBBD: {Code: CurrencyBBD, Symbol: "Bds$", Decimals: 2},
	CurrencyBDT: {Code: CurrencyBDT, Symbol: "৳", Decimals: 2},
	CurrencyBGN: {Code: CurrencyBGN, Symbol: "лв", Decimals: 2},
	CurrencyBHD: {Code: CurrencyBHD, Symbol: ".د.ب", Decimals: 3},
	CurrencyBIF: {Code: CurrencyBIF, Symbol: "FBu", Decimals: 0},
	CurrencyBMD: {Code: CurrencyBMD, Symbol: "$", Decimals: 2},
	CurrencyBND: {Code: CurrencyBND, Symbol: "B$", Decimals: 2},
	CurrencyBOB: {Code: CurrencyBOB, Symbol: "Bs", Decimals: 2},
	CurrencyBRL: {Code: CurrencyBRL, Symbol: "R$", Decimals: 2},
	CurrencyBSD: {Code: CurrencyBSD, Symbol: "B$", Decimals: 2},
	CurrencyBTN: {Code: CurrencyBTN, Symbol: "Nu.", Decimals: 2},
	CurrencyBWP: {Code: CurrencyBWP, Symbol: "P", Decimals: 2},
	CurrencyBYN: {Code: CurrencyBYN, Symbol: "Br", Decimals: 2},
	CurrencyBZD: {Code: CurrencyBZD, Symbol: "BZ$", Decimals: 2},
	CurrencyCAD: {Code: CurrencyCAD, Symbol: "C$", Decimals: 2},
	CurrencyCDF: {Code: CurrencyCDF, Symbol: "FC", Decimals: 2},
	CurrencyCHF: {Code: CurrencyCHF, Symbol: "CHF", Decimals: 2},
	CurrencyCLP: {Code: CurrencyCLP, Symbol: "$", Decimals: 0},
	CurrencyCNY: {Code: CurrencyCNY, Symbol: "¥", Decimals: 2},
	CurrencyCOP: {Code: CurrencyCOP, Symbol: "$", Decimals: 2},
	CurrencyCRC: {Code: CurrencyCRC, Symbol: "₡", Decimals: 2},
	CurrencyCUP: {Code: CurrencyCUP, Symbol: "$", Decimals: 2},
	CurrencyCVE: {Code: CurrencyCVE, Symbol: "$", Decimals: 2},
	CurrencyCZK: {Code: CurrencyCZK, Symbol: "Kč", Decimals: 2},
	CurrencyDJF: {Code: CurrencyDJF, Symbol: "Fdj", Decimals: 0},
	CurrencyDKK: {Code: CurrencyDKK, Symbol: "kr", Decimals: 2},
	CurrencyDOP: {Code: CurrencyDOP, Symbol: "RD$", Decimals: 2},
	CurrencyDZD: {Code: CurrencyDZD, Symbol: "د.ج", Decimals: 2},
	CurrencyEGP: {Code: CurrencyEGP, Symbol: "E£", Decimals: 2},
	CurrencyERN: {Code: CurrencyERN, Symbol: "Nfk", Decimals: 2},
	CurrencyETB: {Code: CurrencyETB, Symbol: "Br", Decimals: 2},
	CurrencyEUR: {Code: CurrencyEUR, Symbol: "€", Decimals: 2},
	CurrencyFJD: {Code: CurrencyFJD, Symbol: "FJ$", Decimals: 2},
	CurrencyFKP: {Code: CurrencyFKP, Symbol: "£", Decimals: 2},
	CurrencyGBP: {Code: CurrencyGBP, Symbol: "£", Decimals: 2},
	CurrencyGEL: {Code: CurrencyGEL, Symbol: "₾", Decimals: 2},
	CurrencyGHS: {Code: CurrencyGHS, Symbol: "₵", Decimals: 2},
	CurrencyGIP: {Code: CurrencyGIP, Symbol: "£", Decimals: 2},
	CurrencyGMD: {Code: CurrencyGMD, Symbol: "D", Decimals: 2},
	CurrencyGNF: {Code: CurrencyGNF, Symbol: "FG", Decimals: 0},
	CurrencyGTQ: {Code: CurrencyGTQ, Symbol: "Q", Decimals: 2},
	CurrencyGYD: {Code: CurrencyGYD, Symbol: "G$", Decimals: 2},
	CurrencyHKD: {Code: CurrencyHKD, Symbol: "HK$", Decimals: 2},
	CurrencyHNL: {Code: CurrencyHNL, Symbol: "L", Decimals: 2},
	CurrencyHRK: {Code: CurrencyHRK, Symbol: "kn", Decimals: 2},
	CurrencyHTG: {Code: CurrencyHTG, Symbol: "G", Decimals: 2},
	CurrencyHUF: {Code: CurrencyHUF, Symbol: "Ft", Decimals: 2},
	CurrencyIDR: {Code: CurrencyIDR, Symbol: "Rp", Decimals: 2},
	CurrencyILS: {Code: CurrencyILS, Symbol: "₪", Decimals: 2},
	CurrencyINR: {Code: CurrencyINR, Symbol: "₹", Decimals: 2},
	CurrencyIQD: {Code: CurrencyIQD, Symbol: "ع.د", Decimals: 3},
	CurrencyIRR: {Code: CurrencyIRR, Symbol: "﷼", Decimals: 2},
	CurrencyISK: {Code: CurrencyISK, Symbol: "kr", Decimals: 0},
	CurrencyJMD: {Code: CurrencyJMD, Symbol: "J$", Decimals: 2},
	CurrencyJOD: {Code: CurrencyJOD, Symbol: "د.ا", Decimals: 3},
	CurrencyJPY: {Code: CurrencyJPY, Symbol: "¥", Decimals: 0},
	CurrencyKES: {Code: CurrencyKES, Symbol: "KSh", Decimals: 2},
	CurrencyKGS: {Code: CurrencyKGS, Symbol: "сом", Decimals: 2},
	CurrencyKHR: {Code: CurrencyKHR, Symbol: "៛", Decimals: 2},
	CurrencyKMF: {Code: CurrencyKMF, Symbol: "CF", Decimals: 0},
	CurrencyKPW: {Code: CurrencyKPW, Symbol: "₩", Decimals: 2},
	CurrencyKRW: {Code: CurrencyKRW, Symbol: "₩", Decimals: 0},
	CurrencyKWD: {Code: CurrencyKWD, Symbol: "د.ك", Decimals: 3},
	CurrencyKYD: {Code: CurrencyKYD, Symbol: "CI$", Decimals: 2},
	CurrencyKZT: {Code: CurrencyKZT, Symbol: "₸", Decimals: 2},
	CurrencyLAK: {Code: CurrencyLAK, Symbol: "₭", Decimals: 2},
	CurrencyLBP: {Code: CurrencyLBP, Symbol: "ل.ل", Decimals: 2},
	CurrencyLKR: {Code: CurrencyLKR, Symbol: "Rs", Decimals: 2},
	CurrencyLRD: {Code: CurrencyLRD, Symbol: "L$", Decimals: 2},
	CurrencyLSL: {Code: CurrencyLSL, Symbol: "L", Decimals: 2},
	CurrencyLYD: {Code: CurrencyLYD, Symbol: "ل.د", Decimals: 3},
	CurrencyMAD: {Code: CurrencyMAD, Symbol: "د.م.", Decimals: 2},
	CurrencyMDL: {Code: CurrencyMDL, Symbol: "L", Decimals: 2},
	CurrencyMGA: {Code: CurrencyMGA, Symbol: "Ar", Decimals: 2},
	CurrencyMKD: {Code: CurrencyMKD, Symbol: "ден", Decimals: 2},
	CurrencyMMK: {Code: CurrencyMMK, Symbol: "K", Decimals: 2},
	CurrencyMNT: {Code: CurrencyMNT, Symbol: "₮", Decimals: 2},
	CurrencyMOP: {Code: CurrencyMOP, Symbol: "MOP$", Decimals: 2},
	CurrencyMRU: {Code: CurrencyMRU, Symbol: "UM", Decimals: 2},
	CurrencyMUR: {Code: CurrencyMUR, Symbol: "₨", Decimals: 2},
	CurrencyMVR: {Code: CurrencyMVR, Symbol: "Rf", Decimals: 2},
	CurrencyMWK: {Code: CurrencyMWK, Symbol: "MK", Decimals: 2},
	CurrencyMXN: {Code: CurrencyMXN, Symbol: "Mex$", Decimals: 2},
	CurrencyMYR: {Code: CurrencyMYR, Symbol: "RM", Decimals: 2},
	CurrencyMZN: {Code: CurrencyMZN, Symbol: "MT", Decimals: 2},
	CurrencyNAD: {Code: CurrencyNAD, Symbol: "N$", Decimals: 2},
	CurrencyNGN: {Code: CurrencyNGN, Symbol: "₦", Decimals: 2},
	CurrencyNIO: {Code: CurrencyNIO, Symbol: "C$", Decimals: 2},
	CurrencyNOK: {Code: CurrencyNOK, Symbol: "kr", Decimals: 2},
	CurrencyNPR: {Code: CurrencyNPR, Symbol: "Rs", Decimals: 2},
	CurrencyNZD: {Code: CurrencyNZD, Symbol: "NZ$", Decimals: 2},
	CurrencyOMR: {Code: CurrencyOMR, Symbol: "ر.ع.", Decimals: 3},
	CurrencyPAB: {Code: CurrencyPAB, Symbol: "B/.", Decimals: 2},
	CurrencyPEN: {Code: CurrencyPEN, Symbol: "S/", Decimals: 2},
	CurrencyPGK: {Code: CurrencyPGK, Symbol: "K", Decimals: 2},
	CurrencyPHP: {Code: CurrencyPHP, Symbol: "₱", Decimals: 2},
	CurrencyPKR: {Code: CurrencyPKR, Symbol: "₨", Decimals: 2},
	CurrencyPLN: {Code: CurrencyPLN, Symbol: "zł", Decimals: 2},
	CurrencyPYG: {Code: CurrencyPYG, Symbol: "₲", Decimals: 0},
	CurrencyQAR: {Code: CurrencyQAR, Symbol: "ر.ق", Decimals: 2},
	CurrencyRON: {Code: CurrencyRON, Symbol: "lei", Decimals: 2},
	CurrencyRSD: {Code: CurrencyRSD, Symbol: "дин", Decimals: 2},
	CurrencyRUB: {Code: CurrencyRUB, Symbol: "₽", Decimals: 2},
	CurrencyRWF: {Code: CurrencyRWF, Symbol: "FRw", Decimals: 0},
	CurrencySAR: {Code: CurrencySAR, Symbol: "ر.س", Decimals: 2},
	CurrencySBD: {Code: CurrencySBD, Symbol: "SI$", Decimals: 2},
	CurrencySCR: {Code: CurrencySCR, Symbol: "₨", Decimals: 2},
	CurrencySDG: {Code: CurrencySDG, Symbol: "ج.س", Decimals: 2},
	CurrencySEK: {Code: CurrencySEK, Symbol: "kr", Decimals: 2},
	CurrencySGD: {Code: CurrencySGD, Symbol: "S$", Decimals: 2},
	CurrencySHP: {Code: CurrencySHP, Symbol: "£", Decimals: 2},
	CurrencySLE: {Code: CurrencySLE, Symbol: "Le", Decimals: 2},
	CurrencySOS: {Code: CurrencySOS, Symbol: "Sh", Decimals: 2},
	CurrencySRD: {Code: CurrencySRD, Symbol: "$", Decimals: 2},
	CurrencySSP: {Code: CurrencySSP, Symbol: "£", Decimals: 2},
	CurrencySTN: {Code: CurrencySTN, Symbol: "Db", Decimals: 2},
	CurrencySYP: {Code: CurrencySYP, Symbol: "£S", Decimals: 2},
	CurrencySZL: {Code: CurrencySZL, Symbol: "L", Decimals: 2},
	CurrencyTHB: {Code: CurrencyTHB, Symbol: "฿", Decimals: 2},
	CurrencyTJS: {Code: CurrencyTJS, Symbol: "ЅМ", Decimals: 2},
	CurrencyTMT: {Code: CurrencyTMT, Symbol: "m", Decimals: 2},
	CurrencyTND: {Code: CurrencyTND, Symbol: "د.ت", Decimals: 3},
	CurrencyTOP: {Code: CurrencyTOP, Symbol: "T$", Decimals: 2},
	CurrencyTRY: {Code: CurrencyTRY, Symbol: "₺", Decimals: 2},
	CurrencyTTD: {Code: CurrencyTTD, Symbol: "TT$", Decimals: 2},
	CurrencyTWD: {Code: CurrencyTWD, Symbol: "NT$", Decimals: 2},
	CurrencyTZS: {Code: CurrencyTZS, Symbol: "TSh", Decimals: 2},
	CurrencyUAH: {Code: CurrencyUAH, Symbol: "₴", Decimals: 2},
	CurrencyUGX: {Code: CurrencyUGX, Symbol: "USh", Decimals: 0},
	CurrencyUSD: {Code: CurrencyUSD, Symbol: "$", Decimals: 2},
	CurrencyUYU: {Code: CurrencyUYU, Symbol: "$U", Decimals: 2},
	CurrencyUZS: {Code: CurrencyUZS, Symbol: "so'm", Decimals: 2},
	CurrencyVES: {Code: CurrencyVES, Symbol: "Bs.S", Decimals: 2},
	CurrencyVND: {Code: CurrencyVND, Symbol: "₫", Decimals: 0},
	CurrencyVUV: {Code: CurrencyVUV, Symbol: "VT", Decimals: 0},
	CurrencyWST: {Code: CurrencyWST, Symbol: "WS$", Decimals: 2},
	CurrencyXAF: {Code: CurrencyXAF, Symbol: "FCFA", Decimals: 0},
	CurrencyXCD: {Code: CurrencyXCD, Symbol: "EC$", Decimals: 2},
	CurrencyXOF: {Code: CurrencyXOF, Symbol: "CFA", Decimals: 0},
	CurrencyXPF: {Code: CurrencyXPF, Symbol: "₣", Decimals: 0},
	CurrencyYER: {Code: CurrencyYER, Symbol: "﷼", Decimals: 2},
	CurrencyZAR: {Code: CurrencyZAR, Symbol: "R", Decimals: 2},
	CurrencyZMW: {Code: CurrencyZMW, Symbol: "ZK", Decimals: 2},
	CurrencyZWL: {Code: CurrencyZWL, Symbol: "Z$", Decimals: 2},
}

func NewCurrency(code CurrencyCode) (Currency, error) {
	currency, ok := Currencies[code]
	if !ok {
		return Currency{}, domainerrors.ErrUnknownCurrencyCode.New(string(code))
	}

	return currency, nil
}

func AllCurrencies() []Currency {
	currencies := make([]Currency, 0, len(Currencies))
	for _, currency := range Currencies {
		currencies = append(currencies, currency)
	}

	return currencies
}

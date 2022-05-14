package printer

import (
	"fmt"
	"net"
	"strings"

	"github.com/xgadget-lab/nexttrace/ipgeo"
	"github.com/xgadget-lab/nexttrace/methods"
)

var dataOrigin string

func TraceroutePrinter(ip net.IP, res map[uint16][]methods.TracerouteHop, dataOrigin string, rdnsenable bool) {
	for hi := uint16(1); hi < 30; hi++ {
		fmt.Print(hi)
		for _, v := range res[hi] {
			hopPrinter(v, rdnsenable)
			if v.Address != nil && ip.String() == v.Address.String() {
				hi = 31
			}
		}
	}
}

func hopPrinter(v2 methods.TracerouteHop, rdnsenable bool) {
	if v2.Address == nil {
		fmt.Println("\t*")
	} else {
		var iPGeoData *ipgeo.IPGeoData
		var err error

		ipStr := v2.Address.String()

		// 判断 err 返回，并且在CLI终端提示错误
		switch dataOrigin {
		case "LeoMoeAPI":
			iPGeoData, err = ipgeo.LeoIP(ipStr)
		case "IP.SB":
			iPGeoData, err = ipgeo.IPSB(ipStr)
		case "IPInfo":
			iPGeoData, err = ipgeo.IPInfo(ipStr)
		case "IPInsight":
			iPGeoData, err = ipgeo.IPInSight(ipStr)
		default:
			iPGeoData, err = ipgeo.LeoIP(ipStr)
		}

		geo := ""
		if err != nil {
			geo = fmt.Sprint("Error: ", err)
		} else {
			geo = formatIpGeoData(ipStr, iPGeoData)
		}

		txt := "\t"

		if rdnsenable {
			ptr, err := net.LookupAddr(ipStr)
			if err != nil {
				txt += fmt.Sprint(ipStr, " ", fmt.Sprintf("%.2f", v2.RTT.Seconds()*1000), "ms ", geo)
			} else {
				txt += fmt.Sprint(ptr[0], " (", ipStr, ") ", fmt.Sprintf("%.2f", v2.RTT.Seconds()*1000), "ms ", geo)
			}
		} else {
			txt += fmt.Sprint(ipStr, " ", fmt.Sprintf("%.2f", v2.RTT.Seconds()*1000), "ms ", geo)
		}

		fmt.Println(txt)
	}
}

func formatIpGeoData(ip string, data *ipgeo.IPGeoData) string {
	var res = make([]string, 0, 10)

	if data.Asnumber == "" {
		res = append(res, "*")
	} else {
		res = append(res, "AS"+data.Asnumber)
	}

	// TODO: 判断阿里云和腾讯云内网，数据不足，有待进一步完善
	if strings.HasPrefix(ip, "9.") {
		res = append(res, "局域网", "腾讯云")
	} else if strings.HasPrefix(ip, "11.") {
		res = append(res, "局域网", "阿里云")
	} else if data.Country == "" {
		res = append(res, "局域网")
	} else {
		// 有些IP的归属信息为空，这个时候将ISP的信息填入
		if data.Owner == "" {
			data.Owner = data.Isp
		}
		if data.District != "" {
			data.City = data.City + ", " + data.District
		}
		if data.Prov == "" && data.City == "" {
			// anyCast或是骨干网数据不应该有国家信息
			data.Owner = data.Owner + ", " + data.Owner
		} else {
			// 非骨干网正常填入IP的国家信息数据
			res = append(res, data.Country)
		}

		if data.Prov != "" {
			res = append(res, data.Prov)
		}
		if data.City != "" {
			res = append(res, data.City)
		}

		if data.Owner != "" {
			res = append(res, data.Owner)
		}
	}

	return strings.Join(res, ", ")
}

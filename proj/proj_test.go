package proj

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

func TestAddDef(t *testing.T) {
	err := addDef("EPSG:102018", "+proj=gnom +lat_0=90 +lon_0=0 +x_0=6300000 +y_0=6300000 +ellps=WGS84 +datum=WGS84 +units=m +no_defs")
	if err != nil {
		t.Error(err)
	}
	err = addDef("testmerc", "+proj=merc +lon_0=5.937 +lat_ts=45.027 +ellps=sphere +datum=none")
	if err != nil {
		t.Error(err)
	}
	err = addDef("testmerc2", "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +units=m +k=1.0 +nadgrids=@null +no_defs")
	if err != nil {
		t.Error(err)
	}
	err = addDef("esriOnline", `PROJCS["WGS_1984_Web_Mercator_Auxiliary_Sphere",GEOGCS["GCS_WGS_1984",DATUM["D_WGS_1984",SPHEROID["WGS_1984",6378137.0,298.257223563]],PRIMEM["Greenwich",0.0],UNIT["Degree",0.0174532925199433]],PROJECTION["Mercator_Auxiliary_Sphere"],PARAMETER["False_Easting",0.0],PARAMETER["False_Northing",0.0],PARAMETER["Central_Meridian",0.0],PARAMETER["Standard_Parallel_1",0.0],PARAMETER["Auxiliary_Sphere_Type",0.0],UNIT["Meter",1.0]]`)
	if err != nil {
		t.Error(err)
	}
}

func TestUnits(t *testing.T) {
	if defs["testmerc2"].Units != "m" {
		t.Error("should parse units")
	}
}

func closeTo(t *testing.T, a, b, tol float64, prefix string) bool {
	if a == math.NaN() || b == math.NaN() || 2*math.Abs(a-b)/math.Abs(a+b) > tol {
		t.Errorf("%s: value should be %f but is %f", prefix, b, a)
		return false
	}
	return true
}

func TestProj2Proj(t *testing.T) {
	// transforming from one projection to another
	sweref99tm, err := Parse("+proj=utm +zone=33 +ellps=GRS80 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs")
	if err != nil {
		t.Error(err)
	}
	rt90, err := Parse("+lon_0=15.808277777799999 +lat_0=0.0 +k=1.0 +x_0=1500000.0 +y_0=0.0 +proj=tmerc +ellps=bessel +units=m +towgs84=414.1,41.3,603.1,-0.855,2.141,-7.023,0 +no_defs")
	if err != nil {
		t.Error(err)
	}
	trans, err := sweref99tm.NewTransform(rt90)
	if err != nil {
		t.Error(err)
	}
	rsltx, rslty, err := trans(319180, 6399862)
	if err != nil {
		t.Error(err)
	}
	closeTo(t, rsltx, 1271137.927154, 0.000001, "x")
	closeTo(t, rslty, 6404230.291456, 0.000001, "y")
}

func TestProj4(t *testing.T) {
	type testPoint struct {
		Code   string
		XY, LL []float64
		Acc    struct {
			XY, LL float64
		}
	}
	var testPoints []testPoint
	f, err := os.Open("testData.json")
	if err != nil {
		t.Fatal(err)
	}
	d := json.NewDecoder(f)
	err = d.Decode(&testPoints)
	if err != nil {
		t.Fatal(err)
	}
	for i, testPoint := range testPoints {
		//if !(i == 52 || i == 53) {
		//	continue
		//}
		if !(strings.Contains(strings.ToLower(testPoint.Code), "merc") ||
			strings.Contains(strings.ToLower(testPoint.Code), "albers") ||
			strings.Contains(strings.ToLower(testPoint.Code), "aea") ||
			strings.Contains(strings.ToLower(testPoint.Code), "lambert") ||
			strings.Contains(strings.ToLower(testPoint.Code), "lcc") ||
			strings.Contains(strings.ToLower(testPoint.Code), "equidistant_conic") ||
			strings.Contains(strings.ToLower(testPoint.Code), "eqdc")) ||
			strings.Contains(strings.ToLower(testPoint.Code), "oblique") ||
			strings.Contains(strings.ToLower(testPoint.Code), "azimuthal") {
			continue
		}

		wgs84, err := Parse("WGS84")
		if err != nil {
			t.Fatal(err)
		}
		xyAcc := 2.
		llAcc := 6.
		if testPoint.Acc.XY != 0 {
			xyAcc = testPoint.Acc.XY
		}
		if testPoint.Acc.LL != 0 {
			llAcc = testPoint.Acc.LL
		}
		xyEPSLN := math.Pow(10, -1*xyAcc)
		llEPSLN := math.Pow(10, -1*llAcc)
		proj, err := Parse(testPoint.Code)
		if err != nil {
			t.Errorf("%s: %s", testPoint.Code, err)
			continue
		}
		trans, err := wgs84.NewTransform(proj)
		if err != nil {
			t.Errorf("%s: %s", testPoint.Code, err)
			continue
		}
		x, y, err := trans(testPoint.LL[0], testPoint.LL[1])
		if err != nil {
			t.Errorf("%d: %s: %s", i, testPoint.Code, err)
			continue
		}
		if !closeTo(t, x, testPoint.XY[0], xyEPSLN, fmt.Sprintf("%s fwd x", testPoint.Code)) {
			continue
		}
		if !closeTo(t, y, testPoint.XY[1], xyEPSLN, fmt.Sprintf("%s fwd y", testPoint.Code)) {
			continue
		}
		trans, err = proj.NewTransform(wgs84)
		if err != nil {
			t.Errorf("%s: %s", testPoint.Code, err)
			continue
		}
		lon, lat, err := trans(testPoint.XY[0], testPoint.XY[1])
		if err != nil {
			t.Errorf("%s: %s", testPoint.Code, err)
			continue
		}
		if !closeTo(t, lon, testPoint.LL[0], llEPSLN, fmt.Sprintf("%d %s inv x", i, testPoint.Code)) {
			continue
		}
		if !closeTo(t, lat, testPoint.LL[1], llEPSLN, fmt.Sprintf("%d %s inv y", i, testPoint.Code)) {
			continue
		}
		//t.Logf("passed %s", testPoint.Code)
	}
}

func TestWKT(t *testing.T) {
	err := addDef("EPSG:4269", `GEOGCS["NAD83",DATUM["North_American_Datum_1983",SPHEROID["GRS 1980",6378137,298.257222101,AUTHORITY["EPSG","7019"]],AUTHORITY["EPSG","6269"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.01745329251994328,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4269"]]`)
	if err != nil {
		t.Fatal(err)
	}
	if defs["EPSG:4269"].ToMeter != 6378137*0.01745329251994328 {
		t.Errorf("want the correct conversion factor (%g) for WKT GEOGCS projections; got %g", 6378137*0.01745329251994328, defs["EPSG:4269"].ToMeter)
	}

	err = addDef("EPSG:4279", `GEOGCS["OS(SN)80",DATUM["OS_SN_1980",SPHEROID["Airy 1830",6377563.396,299.3249646,AUTHORITY["EPSG","7001"]],AUTHORITY["EPSG","6279"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.01745329251994328,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4279"]]`)
	if err != nil {
		t.Fatal(err)
	}
	if defs["EPSG:4279"].ToMeter != 6377563.396*0.01745329251994328 {
		t.Errorf("want the correct conversion factor (%g) for WKT GEOGCS projections; got %g", 6377563.396*0.01745329251994328, defs["EPSG:4279"].ToMeter)
	}
}

func TestErrors(t *testing.T) {
	_, err := Parse("fake one")
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("should throw an error for an unknown ref")
	}
}

func TestDatum(t *testing.T) {
	err := addDef("EPSG:5514", "+proj=krovak +lat_0=49.5 +lon_0=24.83333333333333 +alpha=30.28813972222222 +k=0.9999 +x_0=0 +y_0=0 +ellps=bessel +pm=greenwich +units=m +no_defs +towgs84=570.8,85.7,462.8,4.998,1.587,5.261,3.56")
	if err != nil {
		t.Fatal(err)
	}
	wgs84, err := Parse("WGS84")
	if err != nil {
		t.Fatal(err)
	}
	to, err := Parse("EPSG:5514")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := wgs84.NewTransform(to)
	if err != nil {
		t.Fatal(err)
	}
	x, y, err := trans(12.806988, 49.452262)
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, x, -868208.6070936776, 1.e-8, "Longitude of point from WGS84")
	closeTo(t, y, -1095793.6411470256, 1.e-9, "Latitude of point from WGS84")
	trans2, err := wgs84.NewTransform(to)
	if err != nil {
		t.Fatal(err)
	}
	x2, y2, err := trans2(12.806988, 49.452262)
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, x2, -868208.6070936776, 1.e-8, "Longitude 2nd of point from WGS84")
	closeTo(t, y2, -1095793.6411470256, 1.e-9, "Latitude of 2nd point from WGS84")
}

func TestWKTParse(t *testing.T) {
	wkt := `GEOGCS["NAD83",DATUM["North_American_Datum_1983",SPHEROID["GRS 1980",6378137,298.257222101,AUTHORITY["EPSG","7019"]],TOWGS84[0,0,0,0,0,0,0],AUTHORITY["EPSG","6269"]],PRIMEM["Greenwich",0,AUTHORITY["EPSG","8901"]],UNIT["degree",0.0174532925199433,AUTHORITY["EPSG","9122"]],AUTHORITY["EPSG","4269"]]`

	sr, err := Parse(wkt)
	if err != nil {
		t.Fatal(err)
	}

	want := &SR{
		Name:          "longlat",
		Title:         "",
		SRSCode:       "",
		DatumCode:     "north_american_datum_1983",
		Rf:            298.257222101,
		Lat0:          math.NaN(),
		Lat1:          math.NaN(),
		Lat2:          math.NaN(),
		LatTS:         math.NaN(),
		Long0:         math.NaN(),
		Long1:         math.NaN(),
		Long2:         math.NaN(),
		LongC:         math.NaN(),
		Alpha:         math.NaN(),
		X0:            math.NaN(),
		Y0:            math.NaN(),
		K0:            1,
		K:             math.NaN(),
		A:             6.378137e+06,
		A2:            4.0680631590769e+13,
		B:             6.356752314140356e+06,
		B2:            4.040829998332877e+13,
		Ra:            false,
		Zone:          math.NaN(),
		UTMSouth:      false,
		DatumParams:   []float64{0, 0, 0, 0, 0, 0, 0},
		ToMeter:       111319.4907932736,
		Units:         "degree",
		FromGreenwich: math.NaN(),
		NADGrids:      "",
		Axis:          "enu",
		local:         false,
		sphere:        false,
		Ellps:         "GRS 1980",
		EllipseName:   "",
		Es:            0.006694380022900686,
		E:             0.08181919104281517,
		Ep2:           0.006739496775478856,
		DatumName:     "",
		NoDefs:        false,
		datum: &datum{
			datum_type:   4,
			datum_params: []float64{0, 0, 0, 0, 0, 0, 0},
			a:            6.378137e+06,
			b:            6.356752314140356e+06,
			es:           0.006694380022900686,
			ep2:          0.006739496775478856,
			nadGrids:     ""},
		Czech: false,
	}

	if !sr.Equal(want, 0) {
		t.Errorf("have\n\t%#v\nwant\n\t%#v", sr, want)
	}
}

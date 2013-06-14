package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-hep/croot"
	"github.com/sbinet/go-arff"
)

var (
	fname = flag.String("i", "", "input ARFF file to convert from")
	oname = flag.String("o", "", "output ROOT file to convert into")
)

func main() {
	flag.Parse()

	if *fname == "" {
		if len(os.Args) > 1 {
			*fname = os.Args[1]
		}
	}

	if *oname == "" {
		if len(os.Args) > 2 {
			*oname = os.Args[2]
		}
	}

	if *oname == "" || *fname == "" {
		fmt.Fprintf(
			os.Stderr,
			"**error** you need to give an input file name and an output file name\n",
		)
		flag.Usage()
		os.Exit(1)
	}

	var f io.Reader

	ff, err := os.Open(*fname)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"**error** %v\n",
			err,
		)
		os.Exit(1)
	}
	defer ff.Close()
	f = ff
	if strings.HasSuffix(*fname, ".gz") {
		f, err = gzip.NewReader(ff)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"**error** %v\n",
				err,
			)
			os.Exit(1)
		}
	}

	dec, err := arff.NewDecoder(f)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"**error** %v\n",
			err,
		)
		os.Exit(1)
	}

	o, err := croot.OpenFile(*oname, "recreate", "ARFF event file", 1, 0)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"**error** %v\n",
			err,
		)
		os.Exit(1)
	}
	defer o.Close("")

	fmt.Printf(":: arff file - relation: [%v]\n", dec.Header.Relation)

	tree := croot.NewTree(dec.Header.Relation, dec.Header.Relation, 32)
	if tree == nil {
		err = fmt.Errorf(
			"arff2root: could not create output tree [%s] in file [%s]",
			dec.Header.Relation,
			*oname,
		)
		fmt.Fprintf(
			os.Stderr,
			"**error** %v\n",
			err,
		)
		os.Exit(1)
	}

	type RootVar struct {
		Value interface{}
	}
	idata := make(map[string]interface{})
	//odata := make(map[string]*RootVar)
	odata := make([]interface{}, len(dec.Header.Attrs))
	const bufsize = 32000

	// declare branches
	for i, attr := range dec.Header.Attrs {
		//fmt.Printf(">>> [%d]: %v\n", i, attr)
		switch attr.Type {
		case arff.Integer:
			vv := int64(0)
			br_name := attr.Name
			br_type := br_name + "/L"
			odata[i] = &vv
			_, err = tree.Branch2(br_name, odata[i], br_type, bufsize)
		case arff.Real, arff.Numeric:
			vv := float64(-999.0)
			br_name := attr.Name
			br_type := br_name + "/D"
			odata[i] = &vv
			_, err = tree.Branch2(br_name, odata[i], br_type, bufsize)
		case arff.Nominal:
			vv := ""
			br_name := attr.Name
			br_type := br_name + "/C"
			odata[i] = &vv
			_, err = tree.Branch2(br_name, odata[i], br_type, bufsize)
		default:
			fmt.Fprintf(
				os.Stderr,
				"**error** invalid type for attribute: %v\n",
				attr,
			)
			os.Exit(1)
		}
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"**error** setting up branch for attribute [%s]: %v\n",
				attr.Name,
				err,
			)
			os.Exit(1)
		}
	}

	for {
		err = dec.Decode(idata)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"**error** %v\n",
				err,
			)
			os.Exit(1)
		}
		for i, attr := range dec.Header.Attrs {
			k := attr.Name
			v := idata[k]
			//fmt.Printf(">>> [%v]: %v\n", k, v)
			if v == nil {
				err = fmt.Errorf(
					"nil value for attribute [%v]",
					k,
				)
				fmt.Fprintf(
					os.Stderr,
					"**error** %v\n",
					err,
				)
				os.Exit(1)
			}
			switch attr.Type {
			case arff.Integer:
				vv := odata[i].(*int64)
				*vv = v.(int64)
				//fmt.Printf(">> [%v]: %v %v %T\n", k, v, odata[i], odata[i])
			case arff.Numeric, arff.Real:
				vv := odata[i].(*float64)
				*vv = v.(float64)
				//fmt.Printf(">> [%v]: %v %v %T\n", k, v, odata[i], odata[i])
			case arff.Nominal:
				vv := odata[i].(*string)
				*vv = v.(string)
			}
		}
		_, err = tree.Fill()
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"**error** filling tree: %v\n",
				err,
			)
			os.Exit(1)
		}
	}
	tree.Write("", 0, 0)

}

// EOF

package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Point struct {
	X, Y float64
}

func Touches(points map[int]Point, toTest Point, radius float64) bool {
	for _, p := range points {
		distX := p.X - toTest.X
		distY := p.Y - toTest.Y
		dist := math.Sqrt(distX*distX + distY*distY)
		if dist < radius {
			return true
		}
	}
	return false
}

type Options struct {
	File          string   `short:"f" long:"file" description:"File to output" required:"true"`
	FamilyAndSize []string `short:"t" long:"family-and-size" description:"Families and size to use. format: 'name:size:begin-end'"`
	ColumnNumber  int      `long:"column-number" description:"Number of column to display multiple families" default:"0"`
	TagBorder     float64  `long:"individual-tag-border" description:"border between tags in column layout" default:"0.2"`
	FamilyMargin  float64  `long:"family-margin" description:"margin between families in mm" default:"2.0"`
	ArenaNumber   int      `long:"arena-number" description:"Number of tags to display in an arena" default:"0"`
	Width         float64  `short:"W" long:"width" description:"Width to use" default:"210"`
	Height        float64  `short:"H" long:"height" description:"Height to use" default:"297"`
	PaperBorder   float64  `long:"arena-border" description:"Draw a border" default:"20.0"`
	DPI           int      `short:"d" long:"dpi" description:"DPI to use" default:"2400"`
}

type FamilyAndSize struct {
	Family     *TagFamily
	Size       float64
	Begin, End int
}

func ExtractFamilyAndSizes(list []string) ([]FamilyAndSize, error) {
	res := []FamilyAndSize{}
	for _, fAndSize := range list {
		fargs := strings.Split(fAndSize, ":")
		if len(fargs) <= 1 {
			return res, fmt.Errorf("invalid family specification '%s': need at list family and size in the form '<name>:<size>'", fAndSize)
		}
		if len(fargs) > 3 {
			return res, fmt.Errorf("invalid family specification '%s':  expected '<name>:<size>:<range>'", fAndSize)
		}
		tf, err := GetFamily(fargs[0])
		if err != nil {
			return res, err
		}
		s, err := strconv.ParseFloat(fargs[1], 64)
		if err != nil {
			return res, err
		}

		if len(fargs) == 2 {
			res = append(res, FamilyAndSize{
				Family: tf,
				Size:   s,
				Begin:  0,
				End:    len(tf.Codes),
			})
			continue
		}

		ranges := strings.Split(fargs[2], "-")
		begin := -1
		end := -1
		if len(ranges) > 2 {
			return res, fmt.Errorf("Only supports ranges XX XX- -XX XX-YY, got '%s'", fargs[2])
		}
		if len(ranges) == 1 {
			idx, err := strconv.ParseInt(ranges[0], 10, 64)
			if err != nil {
				return res, err
			}
			begin = int(idx)
			end = int(idx) + 1
		}
		if len(ranges[0]) == 0 {
			begin = 0
		} else {
			idx, err := strconv.ParseInt(ranges[0], 10, 64)
			if err != nil {
				return res, err
			}
			if int(idx) >= len(tf.Codes) {
				return res, fmt.Errorf("%d is out-of-range in %s (size:%d)'", idx, fargs[0], len(tf.Codes))
			}
			begin = int(idx)
		}
		if len(ranges[1]) == 0 {
			end = len(tf.Codes)
		} else {
			idx, err := strconv.ParseInt(ranges[0], 10, 64)
			if err != nil {
				return res, err
			}
			if int(idx) >= len(tf.Codes) {
				return res, fmt.Errorf("%d is out-of-range in %s (size:%d)'", idx, fargs[0], len(tf.Codes))
			}
			end = int(idx)
		}
		res = append(res, FamilyAndSize{
			Family: tf,
			Size:   s,
			Begin:  begin,
			End:    end,
		})
	}
	return res, nil
}

func Execute() error {
	opts := Options{}
	if _, err := flags.Parse(&opts); err != nil {
		return err
	}

	drawer, err := NewSVGDrawer(opts.File, opts.Width, opts.Height, opts.DPI)
	if err != nil {
		return err
	}
	defer drawer.Close()

	families, err := ExtractFamilyAndSizes(opts.FamilyAndSize)
	if err != nil {
		return err
	}

	var layouter Layouter = nil

	if opts.ArenaNumber != 0 && opts.ColumnNumber == 0 {
		layouter = &ArenaLayouter{
			Border: opts.PaperBorder,
			Number: opts.ArenaNumber,
			Width:  opts.Width,
			Height: opts.Height,
		}
	} else if opts.ColumnNumber != 0 && opts.ArenaNumber == 0 {
		return fmt.Errorf("Column layouter is not yet implemented")
	} else if opts.ColumnNumber != 0 && opts.ArenaNumber != 0 {
		return fmt.Errorf("Please specify either a column or either an arena layout")
	}
	if layouter == nil {
		return fmt.Errorf("Please specify a layout with either --arena-number or -- col-number")
	}

	return layouter.Layout(drawer, families)
}

func main() {
	if err := Execute(); err != nil {
		log.Printf("Unhandled error: %s", err)
		os.Exit(1)
	}

}

package main

import (
    "bufio"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

// ====================================================================================================================
// What are the enumerable parts?
// - Rings:     {{AA, BB, CC, DD}, {AB, BC, CD, AB}, ... }
//      - Description: The picture is composed of 16 rings, each ring is composed of 4 segments. Each ring, starting
//                     from the most inner ring is rotated by -k*pi/8 radians. Each ring is assumed to be representative
//                     of 16 bits.
//      - Cardinality: 16
// - Segments:  {{AA}, {AB}, {AC}, ... }
//      - Description:  The segments that exist within the picture. A segment is a combination of a color and a width.
//                      Each segment is assumed to be representative of 4 bits.
//      - Cardinality:  64 (4 * |Rings|)
//      - Permutations: 16 (|Color| * |Widths|)
// - Colors:    {A, B, C, D}
//      - Description:  The color that a segment is. Each color is assumed to be representative of 2 bits.
//      - Cardinality:  4
// - Widths:    {A, B, C, D}
//      - Description:  The width that a segment is. Each width is assumed to be representative of 2 bits.
//      - Cardinality:  4
// ====================================================================================================================
// - 4   | Different types of width
// - 4   | Different types of color
// - 64  | Number of segments in the image
// - 16  | Rings
// - 16  | Rotations
// - 4   | Bits per segment
// - 16  | Bits per ring
// - 256 | Bits total
// ====================================================================================================================
// Hints:
// - The following hint was given: f_i(x_i, y_{i-1})
//      - "There is logic behind it and it is both hinted in the equation and in the puzzle"
//          - What logic is hinted in the puzzle?
//              - Clockwise Rotation
//              - There are a equal number of rings, rotations, and bits that a ring encodes.
//      - What are the individual pieces of this hint?
//          - Function 'f' which is enumerable.
//          - Variable 'x' which is enumerable.
//          - Variable 'y' which is enumerable.
//      - What is 'i' representative of?
//          - What is the set of 'i'?
//          - Is 'i' a subset of the set of integers?
//      - What is 'x' representative of?
//          - Color
//          - Width
//          - Segment
//          - Rotation
//          - Ring
//      - What is 'y' representative of?
//          - Color
//          - Width
//          - Segment
//          - Rotation
//          - Ring
//      - What is the output of the function?
//          - A single hex character:
//              - 64 hex characters for a private key.
//      - Can you make the assumption that i > 1?
//      - Assume that 'x' and 'y' are sequences:
//          - Assume that 'i belongs to the natural numbers, then it is invalid to say that we are trying to get the
//            0th element of the y sequence unless it wrapped around.
//      - Can you make the assumption that x and y are the same length?
//      - Does the y-1 hint at some sort of initial condition?
//      - Why would the last element of y be unimportant?
//      - Why does the function have two inputs?
//      - Why is 'i' even a subscript of the function at all and not some parameter?
//      - Could x and y just be coordinates within the matrix.
//          - 'x' and 'y' could even represent the entire columns and rows of the matrix.
//      - I do not believe that x and y representing color and width makes much sense unless there was some
//        wrapping operation taking place. It should be assumed that all of the data from the image should
//        be used and not repeated because it exactly represents 256 bits. One off (y-1) increasing sequences
//        leave the last element of the sequence to not be used in any such function calculation. It also
//        seems much too simple.
// ====================================================================================================================
// Segment representation:
// - Each segment is representative of 4 bits. It is unclear though how exactly segments are encoded into their
//   4 bit representation. It goes without saying that trying each combination of the segments representing a
//   specific 4 bit value (between 0 to 15) is 16!, which is not a reasonable amount computation. Assuming that
//   the color and width is either the left two bits or right two bits, respectively, is more reasonable. That
//   leaves 4! * 4! * 2 or 1152 possible combinations for each property to be left or right and each represent
//   a specific unique number.
// ====================================================================================================================
// Matrix representation:
// - The image can be represented as a 16 x 16 bit array, with each segment corresponding to a section of the matrix.
//   This assumes that each segment encodes to the same bit sequence. This gives a structure to the data that could
//   lead to different derivations.
// ====================================================================================================================
// Notes:
// - I make the assumption that all of the data is accounted for. This includes 64 segments which are comprised of
//   two separate properties, color and width. The number of colors and widths are both 4, which leads there to be
//   16 possible permutations of a segment (4 * 4). Log2(16) is equal to 4, so the representation of each segment
//   could be encoded as a hex character. Then this would result in 256 bits of data, which is equal to the private
//   key length. This suggests that the data then needs to be properly ordered in some fashion.
// ====================================================================================================================

// On the selected colors, I propose that the values were selected to represent a right rotation
// (clockwise). The reason I believe this is because the bytes appear in ascending order in their
// largest to smallest values and this is the only time they ever fit such a pattern. A descending
// pattern is not possible with this configuration. This proposal is also backed by the fact that
// rings are rotated clockwise (left to right).
//
// On the selected values, I propose that a difference of multiples of 0x22 were selected to hold the previous
// property of having ascending values. An ascending or descending set of rotation values only has three valid
// enumerations. Therefore, if the color is meant to represent 4 different enumerations one must be synthesized
// similarly but distinctly enough to not throw off the previous pattern. Giving a gap of 0x22 allows for values
// to be offset by 0x11 but still use the similar pattern. The puzzle creator used a similar scheme with the
// previous step, weaving values in such a manner. It's important to note though that the puzzle creator
// probably didn't use an ascending pattern for 'A' because the color would be hard to distinguish between
// 'B' visually. The omitted value in that sequence for an ascending pattern is 0x77.
var colors = map[string]string{
    "rgb(51, 85, 51)":   "A",   // 0x335533 | 0x33 0x55
    "rgb(68, 102, 136)": "B",   // 0x446688 | 0x44 0x66 0x88
    "rgb(102, 136, 68)": "C",   // 0x884466 | 0x44 0x66 0x88
    "rgb(136, 68, 102)": "D",   // 0x668844 | 0x44 0x66 0x88
}

// There is nothing remarkable about the widths except that they are in multiples of 5.
var widths = map[string]string{
    "5":  "A",
    "10": "B",
    "15": "C",
    "20": "D",
}

type SVG struct {
    XMLName xml.Name `xml:"svg"`
    Path    []Path   `xml:"path"`
}

type Path struct {
    XMLName xml.Name `xml:"path"`
    Shape   string   `xml:"d,attr"`
    Style   string   `xml:"style,attr"`
}

func main() {
    infile, err := os.Open("w3Lc0Me_tO-7h3.FuNc710nc0rE-GlHf.svg")
    if err != nil {
        log.Fatal("failed to open input file: ", err)
    }

    bs, _ := ioutil.ReadAll(infile)

    var svg SVG
    if err := xml.Unmarshal(bs, &svg); err != nil {
        log.Fatal("failed to unmarshal SVG file: ", err)
    }

    outfile, err := os.Create("out.txt")
    if err != nil {
        log.Fatal("failed to open output file: ", err)
    }

    w := bufio.NewWriter(outfile)

    for _, p := range svg.Path {
        color, width := parseStyle(p.Style)
        w.WriteString(fmt.Sprintf("%s%s\n", color, width))
    }

    w.Flush()
}

func parseStyle(style string) (string, string) {
    color, width := "U", "U"
    for _, s := range strings.Split(style, ";") {
        if s == "" {
            continue
        }

        s = strings.TrimSpace(s)

        i := strings.IndexRune(s, ':')
        v := strings.TrimSpace(s[i+1:])

        switch s[:i] {
        case "stroke":
            if color != "U" {
                log.Fatal("invalid state, color already set")
            }
            color = colors[v]
        case "stroke-width":
            if width != "U" {
                log.Fatal("invalid state, width already set")
            }
            width = widths[v]
        }
    }

    return color, width
}

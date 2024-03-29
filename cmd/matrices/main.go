package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

var permutations = []map[string]int{
    {"A": 0, "B": 1, "C": 2, "D": 3,}, // A,B,C,D
    {"A": 1, "B": 0, "C": 2, "D": 3,}, // B,A,C,D
    {"A": 1, "B": 2, "C": 0, "D": 3,}, // C,A,B,D
    {"A": 0, "B": 2, "C": 1, "D": 3,}, // A,C,B,D
    {"A": 2, "B": 0, "C": 1, "D": 3,}, // B,C,A,D
    {"A": 2, "B": 1, "C": 0, "D": 3,}, // C,B,A,D
    {"A": 3, "B": 1, "C": 0, "D": 2,}, // C,B,D,A
    {"A": 3, "B": 0, "C": 1, "D": 2,}, // B,C,D,A
    {"A": 3, "B": 2, "C": 1, "D": 0,}, // D,C,B,A
    {"A": 3, "B": 2, "C": 0, "D": 1,}, // C,D,B,A
    {"A": 3, "B": 0, "C": 2, "D": 1,}, // B,D,C,A
    {"A": 3, "B": 1, "C": 2, "D": 0,}, // D,B,C,A
    {"A": 1, "B": 3, "C": 2, "D": 0,}, // D,A,C,B
    {"A": 0, "B": 3, "C": 2, "D": 1,}, // A,D,C,B
    {"A": 2, "B": 3, "C": 0, "D": 1,}, // C,D,A,B
    {"A": 2, "B": 3, "C": 1, "D": 0,}, // D,C,A,B
    {"A": 0, "B": 3, "C": 1, "D": 2,}, // A,C,D,B
    {"A": 1, "B": 3, "C": 0, "D": 2,}, // C,A,D,B
    {"A": 1, "B": 0, "C": 3, "D": 2,}, // B,A,D,C
    {"A": 0, "B": 1, "C": 3, "D": 2,}, // A,B,D,C
    {"A": 2, "B": 1, "C": 3, "D": 0,}, // D,B,A,C
    {"A": 2, "B": 0, "C": 3, "D": 1,}, // B,D,A,C
    {"A": 0, "B": 2, "C": 3, "D": 1,}, // A,D,B,C
    {"A": 1, "B": 2, "C": 3, "D": 0,}, // D,A,B,C
}

type Matrix struct {
    Zeros         int         `json:"zeros"`
    Ones          int         `json:"ones"`
    Ordering      int         `json:"ordering"`
    ColorEncoding int         `json:"color_encoding"`
    WidthEncoding int         `json:"width_encoding"`
    Bits          [16][16]int `json:"bits"`
}

func main() {
    infile, err := os.Open("out.txt")
    if err != nil {
        log.Fatal("failed to open segments file: ", err)
    }

    segments := make([]string, 0, 64)

    scanner := bufio.NewScanner(infile)
    for scanner.Scan() {
        segments = append(segments, scanner.Text())
    }

    matrices := make([]*Matrix, 0)

    // Color << 2 | Width ; Width << 2 | Color
    for order := 0; order < 2; order++ {
        for ce := 0; ce < 24; ce++ {
            for we := 0; we < 24; we++ {
                encodedSegments := make([]int, 0, 64)

                for _, segment := range segments {
                    c := permutations[ce][segment[:1]]
                    w := permutations[we][segment[1:]]

                    var encoded int
                    switch order {
                    case 0:
                        encoded = c<<2 | w
                    case 1:
                        encoded = w<<2 | c
                    }

                    encodedSegments = append(encodedSegments, encoded)
                }

                rotated := make([]uint16, 16)

                for ring := 0; ring < 16; ring++ {
                    for off := 0; off < 16; off++ {
                        // Bit position relative to the ring rotation.
                        pos := ring*16 + off

                        // Grab the segment that the bit is referencing to. This is just the bit position
                        // divided by 4.
                        encoded := encodedSegments[pos>>2]

                        // Starting from left to right, pick the bit from the encoded segment.
                        bit := (encoded >> uint(0x3-(pos&0x3))) & 0x1

                        // Set the bit relative to the position in the ring it would be. Left to right encoding.
                        rotated[ring] |= uint16(bit << (15 - uint((off+ring)%16)))
                    }
                }

                zeros := 0
                for _, v := range rotated {
                    zeros += strings.Count(fmt.Sprintf("%016b\n", v), "0")
                }

                matrix := &Matrix{
                    Zeros:         zeros,
                    Ones:          256 - zeros,
                    Ordering:      order,
                    ColorEncoding: ce,
                    WidthEncoding: we,
                }

                for x, v := range rotated {
                    for i := 0; i < 16; i++ {
                        b := v >> uint(15-i) & 0x1
                        matrix.Bits[x][i] = int(b)
                    }
                }

                matrices = append(matrices, matrix)
            }
        }

        bs, err := json.Marshal(&matrices)
        if err != nil {
            log.Fatal("failed to marshal matrices: ", err)
        }

        if err := ioutil.WriteFile("matrices.json", bs, os.ModePerm); err != nil {
            log.Fatal("failed to write matrices file: ", err)
        }
    }
}

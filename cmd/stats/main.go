package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
)

func main() {
    infile, err := os.Open("out.txt")
    if err != nil {
        log.Panic("failed to open segments file: ", err)
    }

    segments := make([]string, 0, 64)

    scanner := bufio.NewScanner(infile)
    for scanner.Scan() {
        segments = append(segments, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        log.Fatal("error while reading segments", err)
    }

    os.MkdirAll("stats", os.ModePerm)

    writeSegmentFrequencies(segments)
    writeColorFrequencies(segments)
    writeWidthFrequencies(segments)
}

func frequencies(slice []string) map[string]int  {
    m := make(map[string]int, 64)
    for _, s := range slice {
        m[s]++
    }
    return m
}

func writeColorFrequencies(segments []string) {
    outfile, err := os.Create("stats/color.frequency.csv")
    if err != nil {
        log.Fatal("failed to create color frequency file: ", err)
    }

    defer outfile.Close()

    colors := make([]string, 0, 64)

    for _, s := range segments {
        colors = append(colors, s[:1])
    }

    printTable(outfile, frequencies(colors))
}

func writeWidthFrequencies(segments []string) {
    outfile, err := os.Create("stats/width.frequency.csv")
    if err != nil {
        log.Fatal("failed to create width frequency file: ", err)
    }

    defer outfile.Close()

    widths := make([]string, 0, 64)

    for _, s := range segments {
        widths = append(widths, s[1:])
    }

    printTable(outfile, frequencies(widths))
}

func writeSegmentFrequencies(segments []string) {
    outfile, err := os.Create("stats/segment.frequency.csv")
    if err != nil {
        log.Fatal("failed to create segment frequency file: ", err)
    }

    defer outfile.Close()

    printTable(outfile, frequencies(segments))
}

func printTable(outfile *os.File, m map[string]int) {
    w := bufio.NewWriter(outfile)

    for k, v := range m {
        w.WriteString(fmt.Sprintf("%s,%d\n", k, v))
    }

    w.Flush()
}
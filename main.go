package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type File string
type CodeBlock []CodeLine
type BlockName string
type language string

type CodeLine struct {
	text   string
	file   File
	lang   language
	number int
}

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock

var namedBlockRe *regexp.Regexp

var fileBlockRe *regexp.Regexp

var replaceRe *regexp.Regexp

var blockStartRe *regexp.Regexp

// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader, inputfilename string) error {

	scanner := bufio.NewReader(r)
	var err error

	var line CodeLine
	line.file = File(inputfilename)

	var inBlock bool
	var bname BlockName
	var fname File
	var block CodeBlock

	var blockPrefix string

	for {
		line.number++
		line.text, err = scanner.ReadString('\n')
		switch err {
		case io.EOF:
			return nil
		case nil:
			// Nothing special
		default:
			return err
		}

		if inBlock {
			line.text = strings.TrimPrefix(line.text, blockPrefix)
			if line.text == "```\n" || line.text == "~~~\n" {

				inBlock = false
				// Update the files map if it's a file.
				if fname != "" {
					files[fname] = append(files[fname], block...)
				}

				// Update the named block map if it's a named block.
				if bname != "" {
					blocks[bname] = append(blocks[bname], block...)
				}

				continue
			}

			block = append(block, line)

			continue
		}

		if matches := blockStartRe.FindStringSubmatch(line.text); matches != nil {
			inBlock = true
			blockPrefix = matches[1]
			line.text = strings.TrimPrefix(line.text, blockPrefix)
			// We were outside of a block, so just blindly reset it.
			block = make(CodeBlock, 0)

			fname, bname, line.lang = parseHeader(line.text)
			if string(fname) == "" && string(bname) == "" && string(line.lang) != "" {
				fname = File(string(line.file) + "." + string(line.lang))
			}

		}

	}

}

func parseHeader(line string) (File, BlockName, language) {
	line = strings.TrimSpace(line)

	var matches []string
	if matches = namedBlockRe.FindStringSubmatch(line); matches != nil {
		return "", BlockName(matches[2]), language(matches[1])
	}
	if matches = fileBlockRe.FindStringSubmatch(line); matches != nil {
		return File(matches[2]), "", language(matches[1])
	}
	return "", "", ""

}

// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace(prefix string) (ret CodeBlock) {

	var line string
	for _, v := range c {
		line = v.text

		matches := replaceRe.FindStringSubmatch(line)
		if matches == nil {
			if v.text != "\n" {
				v.text = prefix + v.text
			}
			ret = append(ret, v)
			continue
		}

		bname := BlockName(matches[2])
		if val, ok := blocks[bname]; ok {
			ret = append(ret, val.Replace(prefix+matches[1])...)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
			ret = append(ret, v)
		}

	}
	return

}

// Finalize extract the textual lines from CodeBlocks and (if needed) prepend a
// notice about "unexpected" filename or line changes, which is extracted from
// the contained CodeLines. The result is a string with newlines ready to be
// pasted into a file.
func (c CodeBlock) Finalize() (ret string) {
	var file File
	var formatstring string
	var linenumber int
	for _, l := range c {
		if linenumber+1 != l.number || file != l.file {
			switch l.lang {
			case "go", "golang":
				formatstring = "//line %[2]v:%[1]v\n"
			case "C", "c", "cpp":
				formatstring = "#line %v \"%v\"\n"
			default:
				ret += l.text
				continue
			}
			ret += fmt.Sprintf(formatstring, l.number, l.file)
		}
		ret += l.text
		linenumber = l.number
		file = l.file
	}
	return
}

func main() {

	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)

	namedBlockRe = regexp.MustCompile("^(?:```|~~~)\\s?([\\w\\+]*)\\s*\"(.+)\"$")

	fileBlockRe = regexp.MustCompile("^(?:```|~~~)\\s?([\\w\\+]+)\\s+>\\s*([\\w\\.\\-\\/]*)$")

	replaceRe = regexp.MustCompile(`^([\s]*)<<<(.+)>>>[\s]*$`)

	blockStartRe = regexp.MustCompile("^([\\s]*)(?:```|~~~)")

	outdir := flag.String("o", ".", "Output directory. If not specified, output is written to the current directory.")
	flag.Parse()

	for _, file := range flag.Args() {

		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: ", err)
			continue
		}

		if err := ProcessFile(f, file); err != nil {
			fmt.Fprintln(os.Stderr, "error: ", err)
		}
		// Don't defer since we're in a loop, we don't want to wait until the function
		// exits.
		f.Close()

	}

	for filename, codeblock := range files {
		fullpath := filepath.Join(*outdir, string(filename))
		if dir := filepath.Dir(fullpath); dir != "." {
			if err := os.MkdirAll(dir, 0775); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}

		f, err := os.Create(string(fullpath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		fmt.Fprintf(f, "%s", codeblock.Replace("").Finalize())
		// We don't defer this so that it'll get closed before the loop finishes.
		f.Close()
	}

}

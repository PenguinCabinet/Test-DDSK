package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func sjis_to_utf8(str string) string {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(str), japanese.ShiftJIS.NewDecoder()))
	if err != nil {
		return ""
	}
	return string(ret)
}

type exec_output struct {
	output_str string
}

func (o *exec_output) Write(b []byte) (int, error) {
	//fmt.Println(sjis_to_utf8(string(b)))
	o.output_str += (string(b))
	return len(b), nil
}

func Is_success_mes(mes string, v bool) {
	success_color := color.New(color.FgGreen).Add(color.Underline)
	failure_color := color.New(color.FgRed).Add(color.Underline)

	fmt.Print(mes)
	if v {
		success_color.Printf("Success\n")
	} else {
		failure_color.Printf("Failure\n")
	}
}

func str_to_ddsk_seq(v string) []int {
	A := []int{}
	text := []string{"ドド", "スコ", "ラブ注入♡"}
	temp_buf := ""
	for len(v) > 0 {
		is_correct_str := false
		for i, e := range text {
			len_data := len([]rune(e))
			if len([]rune(v)) >= len_data && string([]rune(v)[:len_data]) == e {
				if temp_buf != "" {
					temp_buf = ""
					A = append(A, -1)
				}
				A = append(A, i)
				v = string([]rune(v)[len_data:])
				is_correct_str = true
				break
			}
		}
		if !is_correct_str {
			temp_buf += string([]rune(v)[:1])
			v = string([]rune(v)[1:])
		}
	}
	if temp_buf != "" {
		temp_buf = ""
		A = append(A, -1)
	}
	return A
}

func main() {
	app := &cli.App{
		Name:  "Test-DDSK",
		Usage: "Check if the DDSK program is right",
		Action: func(c *cli.Context) error {
			//fmt.Printf("%v,%v\n", c.Args().First(), c.Args().Slice()[1:])
			out := exec_output{}
			err_out := exec_output{}
			temp := []string{}
			if c.Args().Len() > 1 {
				temp = c.Args().Slice()[1:]
			}
			cmd := exec.Command(c.Args().First(), temp...)
			cmd.Env = append(os.Environ(),
				"LANG=ja_JP.UTF8",
				"PYTHONUTF8=1",
			)
			//stdout, _ := cmd.StdoutPipe()
			cmd.Stdout = &out
			cmd.Stderr = &err_out
			cmd.Run()

			output_str := (out.output_str)
			output_str = strings.Replace(output_str, "\n", "", -1)
			output_str = strings.Replace(output_str, " ", "", -1)
			output_str = strings.ReplaceAll(output_str, "\t", "")
			output_str = strings.ReplaceAll(output_str, "\r", "")

			output_seq := str_to_ddsk_seq(output_str)

			Is_success_mes("エラーメッセージは存在するか ", err_out.output_str == "")
			Is_success_mes("ドドスコスコスコが三回表示された後ラブ注入♡で終了しているか ", (func() bool {
				correct_seq := []int{0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1, 2}
				if len(output_seq) < len(correct_seq) {
					return false
				}

				j := len(correct_seq) - 1
				for i := len(output_seq) - 1; i > len(output_seq)-1-len(correct_seq); i-- {
					if output_seq[i] != correct_seq[j] {
						return false
					}
					j -= 1
				}
				return true
			})())

			Is_success_mes("[ドド,スコ,ラブ注入♡]以外の文字が使用されていないか ", (func() bool {
				for _, e := range output_seq {
					if e == -1 {
						return false
					}
				}
				return true
			})())

			Is_success_mes("ドドスコスコスコが三回表示された後ラブ注入♡が起きているにもかかわらず停止していない ", (func() bool {
				correct_seq := []int{0, 1, 1, 1, 0, 1, 1, 1, 0, 1, 1, 1, 2}
				C := 0
				for i2 := 0; ; i2++ {
					if len(output_seq)-i2 < len(correct_seq) {
						break
					}

					Is_add_C := true
					j := len(correct_seq) - 1
					for i := len(output_seq) - 1 - i2; i > len(output_seq)-1-len(correct_seq)-i2; i-- {
						if output_seq[i] != correct_seq[j] {
							Is_add_C = false
						}
						j -= 1
					}
					if Is_add_C {
						C += 1
					}
				}
				return C <= 1
			})())

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

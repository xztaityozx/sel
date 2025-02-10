package test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runSel(sel string, args, stdin []string) (stdout, stderr []string, err error) {
	newLine := "\n"
	if runtime.GOOS == "windows" {
		newLine = "\r\n"
	}

	cmd := exec.Command(sel, args...)

	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	cmd.Stdin = strings.NewReader(strings.Join(stdin, newLine))

	err = cmd.Run()
	if err != nil {
		return nil, nil, err
	}

	stdout = strings.Split(strings.TrimRight(stdoutBuffer.String(), newLine), newLine)
	stderr = strings.Split(strings.TrimRight(stderrBuffer.String(), newLine), newLine)

	return
}

func Test_E2E(t *testing.T) {
	as := assert.New(t)
	selPath := filepath.Join(ProjectRoot(), "dist", "sel")

	type input struct {
		args  []string
		stdin []string
	}

	testcases := []struct {
		name           string
		input          input
		expectedStdout []string
		expectedStderr []string
		expectedError  error
	}{
		{
			name: "sel 1 to be 1..10",
			input: input{
				args:  []string{"1"},
				stdin: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
			expectedStdout: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel 2 to be 2, 4, 6, 8, 10...20",
			input: input{
				args:  []string{"2"},
				stdin: []string{"1 2", "3 4", "5 6", "7 8", "9 10", "11 12", "13 14", "15 16", "17 18", "19 20"},
			},
			expectedStdout: []string{"2", "4", "6", "8", "10", "12", "14", "16", "18", "20"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel 1:4 to be 1 2 3 4, 11 12 13 14 ... 91 92 93 94",
			input: input{
				args: []string{"1:4"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10",
					"11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30",
					"31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50",
					"51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70",
					"71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90",
					"91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{
				"1 2 3 4",
				"11 12 13 14",
				"21 22 23 24",
				"31 32 33 34",
				"41 42 43 44",
				"51 52 53 54",
				"61 62 63 64",
				"71 72 73 74",
				"81 82 83 84",
				"91 92 93 94",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel 1 2 3 4 to be 1 2 3 4, 11 12 13 14 ... 91 92 93 94",
			input: input{
				args: []string{"1", "2", "3", "4"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10",
					"11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30",
					"31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50",
					"51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70",
					"71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90",
					"91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{
				"1 2 3 4",
				"11 12 13 14",
				"21 22 23 24",
				"31 32 33 34",
				"41 42 43 44",
				"51 52 53 54",
				"61 62 63 64",
				"71 72 73 74",
				"81 82 83 84",
				"91 92 93 94",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel 0 prints all columns",
			input: input{
				args: []string{"1", "2", "3", "4"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10",
					"11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30",
					"31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50",
					"51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70",
					"71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90",
					"91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{
				"1 2 3 4",
				"11 12 13 14",
				"21 22 23 24",
				"31 32 33 34",
				"41 42 43 44",
				"51 52 53 54",
				"61 62 63 64",
				"71 72 73 74",
				"81 82 83 84",
				"91 92 93 94",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel --template 'template: {}' 1 prints template: 1, template: 2, ...",
			input: input{
				args:  []string{"--template", "template: {}", "1"},
				stdin: []string{"1", "2", "3", "4"},
			},
			expectedStdout: []string{
				"template: 1",
				"template: 2",
				"template: 3",
				"template: 4",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -d , 3 print 3,7,11,...",
			input: input{
				args: []string{"-d", ",", "3"},
				stdin: []string{
					"1,2,3,4",
					"5,6,7,8",
					"9,10,11,12",
					"13,14,15,16",
				},
			},
			expectedStdout: []string{"3", "7", "11", "15"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -d , -D @ 1 2 prints 1@2, 5@6, 9@10, ...",
			input: input{
				args: []string{"-d", ",", "-D", "@", "1", "2"},
				stdin: []string{
					"1,2,3,4",
					"5,6,7,8",
					"9,10,11,12",
					"13,14,15,16",
				},
			},
			expectedStdout: []string{"1@2", "5@6", "9@10", "13@14"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -d , -D @ 0 prints all columns with @",
			input: input{
				args: []string{"-d", ",", "-D", "@", "0"},
				stdin: []string{
					"1,2,3,4",
					"5,6,7,8",
					"9,10,11,12",
					"13,14,15,16",
				},
			},
			expectedStdout: []string{
				"1@2@3@4",
				"5@6@7@8",
				"9@10@11@12",
				"13@14@15@16",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel --use-regexp -d ',|@' 2:8 prints 2 3 4 5 6 7 8, ...",
			input: input{
				args: []string{"--use-regexp", "-d", ",|@", "2:8"},
				stdin: []string{
					"1,2,3,4,5,6,7,8,9,10@1@2@3@4@5@6@7@8@9@10",
					"11,12,13,14,15,16,17,18,19,20@11@12@13@14@15@16@17@18@19@20",
					"21,22,23,24,25,26,27,28,29,30@21@22@23@24@25@26@27@28@29@30",
					"31,32,33,34,35,36,37,38,39,40@31@32@33@34@35@36@37@38@39@40",
					"41,42,43,44,45,46,47,48,49,50@41@42@43@44@45@46@47@48@49@50",
					"51,52,53,54,55,56,57,58,59,60@51@52@53@54@55@56@57@58@59@60",
					"61,62,63,64,65,66,67,68,69,70@61@62@63@64@65@66@67@68@69@70",
					"71,72,73,74,75,76,77,78,79,80@71@72@73@74@75@76@77@78@79@80",
					"81,82,83,84,85,86,87,88,89,90@81@82@83@84@85@86@87@88@89@90",
					"91,92,93,94,95,96,97,98,99,100@91@92@93@94@95@96@97@98@99@100",
				},
			},
			expectedStdout: []string{
				"2 3 4 5 6 7 8",
				"12 13 14 15 16 17 18",
				"22 23 24 25 26 27 28",
				"32 33 34 35 36 37 38",
				"42 43 44 45 46 47 48",
				"52 53 54 55 56 57 58",
				"62 63 64 65 66 67 68",
				"72 73 74 75 76 77 78",
				"82 83 84 85 86 87 88",
				"92 93 94 95 96 97 98",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		}, {
			name: "sel -g -d ',|@' 2:8 prints 2 3 4 5 6 7 8, ...",
			input: input{
				args: []string{"-g", "-d", ",|@", "2:8"},
				stdin: []string{
					"1,2,3,4,5,6,7,8,9,10@1@2@3@4@5@6@7@8@9@10",
					"11,12,13,14,15,16,17,18,19,20@11@12@13@14@15@16@17@18@19@20",
					"21,22,23,24,25,26,27,28,29,30@21@22@23@24@25@26@27@28@29@30",
					"31,32,33,34,35,36,37,38,39,40@31@32@33@34@35@36@37@38@39@40",
					"41,42,43,44,45,46,47,48,49,50@41@42@43@44@45@46@47@48@49@50",
					"51,52,53,54,55,56,57,58,59,60@51@52@53@54@55@56@57@58@59@60",
					"61,62,63,64,65,66,67,68,69,70@61@62@63@64@65@66@67@68@69@70",
					"71,72,73,74,75,76,77,78,79,80@71@72@73@74@75@76@77@78@79@80",
					"81,82,83,84,85,86,87,88,89,90@81@82@83@84@85@86@87@88@89@90",
					"91,92,93,94,95,96,97,98,99,100@91@92@93@94@95@96@97@98@99@100",
				},
			},
			expectedStdout: []string{
				"2 3 4 5 6 7 8",
				"12 13 14 15 16 17 18",
				"22 23 24 25 26 27 28",
				"32 33 34 35 36 37 38",
				"42 43 44 45 46 47 48",
				"52 53 54 55 56 57 58",
				"62 63 64 65 66 67 68",
				"72 73 74 75 76 77 78",
				"82 83 84 85 86 87 88",
				"92 93 94 95 96 97 98",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -gd ',|@' 2:8 prints 2 3 4 5 6 7 8, ...",
			input: input{
				args: []string{"-gd", ",|@", "2:8"},
				stdin: []string{
					"1,2,3,4,5,6,7,8,9,10@1@2@3@4@5@6@7@8@9@10",
					"11,12,13,14,15,16,17,18,19,20@11@12@13@14@15@16@17@18@19@20",
					"21,22,23,24,25,26,27,28,29,30@21@22@23@24@25@26@27@28@29@30",
					"31,32,33,34,35,36,37,38,39,40@31@32@33@34@35@36@37@38@39@40",
					"41,42,43,44,45,46,47,48,49,50@41@42@43@44@45@46@47@48@49@50",
					"51,52,53,54,55,56,57,58,59,60@51@52@53@54@55@56@57@58@59@60",
					"61,62,63,64,65,66,67,68,69,70@61@62@63@64@65@66@67@68@69@70",
					"71,72,73,74,75,76,77,78,79,80@71@72@73@74@75@76@77@78@79@80",
					"81,82,83,84,85,86,87,88,89,90@81@82@83@84@85@86@87@88@89@90",
					"91,92,93,94,95,96,97,98,99,100@91@92@93@94@95@96@97@98@99@100",
				},
			},
			expectedStdout: []string{
				"2 3 4 5 6 7 8",
				"12 13 14 15 16 17 18",
				"22 23 24 25 26 27 28",
				"32 33 34 35 36 37 38",
				"42 43 44 45 46 47 48",
				"52 53 54 55 56 57 58",
				"62 63 64 65 66 67 68",
				"72 73 74 75 76 77 78",
				"82 83 84 85 86 87 88",
				"92 93 94 95 96 97 98",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -d . 2::2 prints パトカータクシー",
			input: input{
				args:  []string{"-d", ".", "2::2"},
				stdin: []string{"パ.タ.ト.ク.カ.シ.ー.ー"},
			},
			expectedStdout: []string{"タ ク シ ー"},
			expectedStderr: []string{""},
			expectedError:  nil,
		}, {
			name: "sel -d '.' -D '-' 1::2 prints パ-ト-カ-ー",
			input: input{
				args:  []string{"-d", ".", "-D", "-", "1::2"},
				stdin: []string{"パ.タ.ト.ク.カ.シ.ー.ー"},
			},
			expectedStdout: []string{"パ-ト-カ-ー"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel 1:20:3 prints 1 4 7 10 13 16 19, ...",
			input: input{
				args: []string{"1:20:3"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{
				"1 4 7 10 13 16 19",
				"21 24 27 30 33 36 39",
				"41 44 47 50 53 56 59",
				"61 64 67 70 73 76 79",
				"81 84 87 90 93 96 99",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel --csv 2 prints 1,2,3",
			input: input{
				args:  []string{"--csv", "2"},
				stdin: []string{"1,\"1,2,3\",3"},
			},
			expectedStdout: []string{"1,2,3"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel 1 2 3 prints 1  2",
			input: input{
				args:  []string{"1", "2", "3"},
				stdin: []string{"1  2  3"},
			},
			expectedStdout: []string{"1  2"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -a 1 2 prints 1 2",
			input: input{
				args:  []string{"-a", "1", "2"},
				stdin: []string{"1  2  3  4"},
			},
			expectedStdout: []string{"1 2"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel --remove-empty 1 2 prints 1 2",
			input: input{
				args:  []string{"--remove-empty", "1", "2"},
				stdin: []string{"1  2  3  4"},
			},
			expectedStdout: []string{"1 2"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -r 1 2 prints 1 2",
			input: input{
				args:  []string{"-r", "1", "2"},
				stdin: []string{"1  2  3  4"},
			},
			expectedStdout: []string{"1 2"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel --tsv 1 2 prints 1 2\t3\t4",
			input: input{
				args:  []string{"--tsv", "1", "2"},
				stdin: []string{"1\t\"2\t3\t4\"\t5"},
			},
			expectedStdout: []string{"1 2\t3\t4"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel /a/:/a/ prints a4 b3 b4 a6 a7 b6 a3 a2 b0 a5 a8 a9 a1 a0",
			input: input{
				args: []string{"/a/:/a/"},
				stdin: []string{
					"b1 b9 b2 a4 b3 b4 a6 b5 b7 a7 b6 a3 a2 b0 a5 b8 a8 a9 a1 a0",
				},
			},
			expectedStdout: []string{
				"a4 b3 b4 a6 a7 b6 a3 a2 b0 a5 a8 a9 a1 a0",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		}, {
			name: "sel 1:/a/ prints b1 b9 b2 a4",
			input: input{
				args: []string{"1:/a/"},
				stdin: []string{
					"b1 b9 b2 a4 b3 b4 a6 b5 b7 a7 b6 a3 a2 b0 a5 b8 a8 a9 a1 a0",
				},
			},
			expectedStdout: []string{
				"b1 b9 b2 a4",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		}, {
			name: "sel /b0/:+3 prints b1 b9 b2 a4",
			input: input{
				args: []string{"/b0/:+3"},
				stdin: []string{
					"b1 b9 b2 a4 b3 b4 a6 b5 b7 a7 b6 a3 a2 b0 a5 b8 a8 a9 a1 a0",
				},
			},
			expectedStdout: []string{
				"b0 a5 b8 a8",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel --split-before 3 prints 3,23,43,63,83",
			input: input{
				args: []string{"--split-before", "3"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{"3", "23", "43", "63", "83"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -S 3 prints 3,23,43,63,83",
			input: input{
				args: []string{"-S", "3"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{"3", "23", "43", "63", "83"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -- -1 prints 20,40,60,80,100",
			input: input{
				args: []string{"--", "-1"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{"20", "40", "60", "80", "100"},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -- -4: prints 17 18 19 20, 37 38 39 40, 57 58 59 60, 77 78 79 80, 97 98 99 100",
			input: input{
				args: []string{"--", "-4:"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{
				"17 18 19 20",
				"37 38 39 40",
				"57 58 59 60",
				"77 78 79 80",
				"97 98 99 100",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
		{
			name: "sel -- -10::2 prints 17 18 19 20, 37 38 39 40, 57 58 59 60, 77 78 79 80, 97 98 99 100",
			input: input{
				args: []string{"--", "-10::2"},
				stdin: []string{
					"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20",
					"21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40",
					"41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57 58 59 60",
					"61 62 63 64 65 66 67 68 69 70 71 72 73 74 75 76 77 78 79 80",
					"81 82 83 84 85 86 87 88 89 90 91 92 93 94 95 96 97 98 99 100",
				},
			},
			expectedStdout: []string{
				"11 13 15 17 19",
				"31 33 35 37 39",
				"51 53 55 57 59",
				"71 73 75 77 79",
				"91 93 95 97 99",
			},
			expectedStderr: []string{""},
			expectedError:  nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			stdout, stderr, err := runSel(selPath, testcase.input.args, testcase.input.stdin)
			if testcase.expectedError != nil {
				as.Equal(err, testcase.expectedError, "エラー内容が一致するべき")
			} else {
				as.Equal(testcase.expectedStdout, stdout, "標準出力が一致するべき")
				as.Equal(testcase.expectedStderr, stderr, "標準エラー出力が一致するべき")
			}
		})
	}
}

// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/hjr265/cactus/cube"

	"github.com/hjr265/cactus/data"
)

func Loop() {
	for {
		func() {
			exec, err := Queue.Pop(true)
			catch(err)
			if exec == nil {
				return
			}
			defer func() {
				if err := recover(); err != nil {
					panic(err)
					log.Print(err)

					exec.Status = 6
					err := exec.Put()
					catch(err)

				} else if exec.Apply {
					subm, err := exec.Submission()
					catch(err)

					subm.Apply(exec)
					err = subm.Put()
					catch(err)
				}
			}()

			log.Printf("Processing Execution:%d", exec.Id)

			exec.Status = 1
			err = exec.Put()
			catch(err)

			subm, err := exec.Submission()
			catch(err)
			prob, err := subm.Problem()
			catch(err)

			runCube, err := cube.New()
			catch(err)
			defer func() {
				err := runCube.Dispose()
				trace(err)
			}()

			var chkCube cube.Cube
			if prob.Checker.Language != "" {
				chkCube, err = cube.New()
				catch(err)

				stack := Stacks[prob.Checker.Language]

				sourceBlob, err := GetBlob(prob.Checker.SourceKey)
				catch(err)
				proc, err := stack.Build(chkCube, sourceBlob)
				catch(err)
				err = sourceBlob.Close()
				catch(err)

				if proc != nil {
					err = chkCube.Execute(proc)
					trace(err)
				}
			}

			stack := Stacks[subm.Language]

			sourceBlob, err := GetBlob(subm.SourceKey)
			catch(err)
			proc, err := stack.Build(runCube, sourceBlob)
			catch(err)
			err = sourceBlob.Close()
			catch(err)

			if proc != nil {
				err = runCube.Execute(proc)
				if err != nil || !proc.Success {
					exec.Build.Error = string(proc.Stderr)

					exec.Verdict = data.CompilationError

					exec.Status = 7
					err = exec.Put()
					catch(err)
					return
				}
			}

			exec.Status = 2
			err = exec.Put()
			catch(err)

			for i, test := range prob.Tests {
				inBlob, err := GetBlob(test.InputKey)
				catch(err)

				proc = stack.Run(runCube)
				catch(err)

				proc.Stdin, err = ioutil.ReadAll(inBlob)
				catch(err)
				err = inBlob.Close()
				catch(err)

				proc.Limits.Cpu = prob.Limits.Cpu
				proc.Limits.Memory = prob.Limits.Memory

				err = runCube.Execute(proc)
				catch(err)

				outKey := fmt.Sprintf("executions:%d:tests:%d:out", exec.Id, i)
				_, err = PutBlob(outKey, bytes.NewReader(proc.Stdout))
				catch(err)

				diff := 0
				pts := 0
				if prob.Checker.Language == "" {
					ansBlob, err := GetBlob(test.AnswerKey)
					catch(err)
					for leftBuf, rightBuf := bufio.NewReader(ansBlob), bufio.NewReader(bytes.NewReader(proc.Stdout)); ; {
						left := []byte{}
						leftEof := false
						for {
							part, pref, err := leftBuf.ReadLine()
							if err == io.EOF {
								leftEof = true
								break
							}
							catch(err)
							left = append(left, part...)
							if !pref {
								break
							}
						}

						right := []byte{}
						rightEof := false
						for {
							part, pref, err := rightBuf.ReadLine()
							if err == io.EOF {
								rightEof = true
								break
							}
							catch(err)
							right = append(right, part...)
							if !pref || len(right) > len(left) {
								break
							}
						}

						if !bytes.Equal(left, right) {
							diff++
							pts = test.Points
						}

						if leftEof && rightEof {
							break
						}
					}

				} else {
					stack := Stacks[prob.Checker.Language]

					chkProc := stack.Run(chkCube)
					catch(err)

					chkProc.Stdin = proc.Stdout

					chkProc.Limits.Cpu = 16
					chkProc.Limits.Memory = 1024

					err = chkCube.Execute(chkProc)
					catch(err)

					n, _ := fmt.Fscanf(bytes.NewReader(chkProc.Stdout), "%d %d", &diff, &pts)
					if n == 0 {
						diff = 1
					}
				}

				verdict := data.Verdict(0)
				switch {
				case proc.Usages.Cpu > prob.Limits.Cpu:
					verdict = data.CpuLimitExceeded
					proc.Usages.Cpu = prob.Limits.Cpu

				case proc.Usages.Memory > prob.Limits.Memory:
					verdict = data.MemoryLimitExceeded
					proc.Usages.Memory = prob.Limits.Memory

				case diff == 0:
					verdict = data.Accepted
					if prob.Checker.Language == "" {
						pts = test.Points
					}

				default:
					verdict = data.WrongAnswer
				}

				exec.Tests = append(exec.Tests, struct {
					OutputKey  string       `json:"outputKey"`
					Difference int          `json:"difference"`
					Verdict    data.Verdict `json:"verdict"`
					Usages     struct {
						Cpu    float64 `json:"cpu"`
						Memory int     `json:"memory"`
					} `json:"usages"`
					Points int `json:"points"`
				}{
					OutputKey:  outKey,
					Difference: diff,
					Verdict:    verdict,
					Usages: struct {
						Cpu    float64 `json:"cpu"`
						Memory int     `json:"memory"`
					}{
						Cpu:    proc.Usages.Cpu,
						Memory: proc.Usages.Memory,
					},
					Points: pts,
				})

				err = exec.Put()
				catch(err)
			}

			exec.Verdict = data.Accepted
			for _, test := range exec.Tests {
				if test.Verdict != data.Accepted {
					exec.Verdict = test.Verdict
					break
				}
			}

			exec.Status = 7
			err = exec.Put()
			catch(err)
		}()
	}
}

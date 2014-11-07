// Copyright 2014 The Cactus Authors. All rights reserved.

package belt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/hjr265/jail.go/jail"

	"github.com/hjr265/cactus/data"
	"github.com/hjr265/cactus/hub"
)

var chNext = make(chan *data.Execution, 4096)

func Push(exec *data.Execution) {
	chNext <- exec
}

func Loop() {
	for {
		func() {
			exec := <-chNext
			defer func() {
				if err := recover(); err != nil {
					log.Print(err)

					exec.Status = 6
					err := exec.Put()
					catch(err)
					hub.Send([]interface{}{"SYNC", "executions", exec.Id})

				} else if exec.Apply {
					subm, err := exec.Submission()
					catch(err)

					subm.Apply(exec)
					err = subm.Put()
					catch(err)
					hub.Send([]interface{}{"SYNC", "submissions", subm.Id})
				}
			}()

			exec.Status = 1
			err := exec.Put()
			catch(err)
			hub.Send([]interface{}{"SYNC", "executions", exec.Id})

			subm, err := exec.Submission()
			catch(err)
			prob, err := subm.Problem()
			catch(err)

			dir, err := ioutil.TempDir("", "")
			catch(err)
			cell := &jail.Cell{
				Dir: dir,
			}
			defer func() {
				err := cell.Dispose()
				trace(err)
			}()

			var chkCell *jail.Cell
			if prob.Checker.Language != "" {
				dir, err := ioutil.TempDir("", "")
				catch(err)
				chkCell = &jail.Cell{
					Dir: dir,
				}

				stack := Stacks[prob.Checker.Language]

				sourceBlob, err := data.Blobs.Get(prob.Checker.SourceKey)
				catch(err)
				cmd, err := stack.Build(chkCell, sourceBlob)
				catch(err)
				err = sourceBlob.Close()
				catch(err)

				if cmd != nil {
					err = cmd.Run()
					trace(err)
				}
			}

			stack := Stacks[subm.Language]

			sourceBlob, err := data.Blobs.Get(subm.SourceKey)
			catch(err)
			cmd, err := stack.Build(cell, sourceBlob)
			catch(err)
			err = sourceBlob.Close()
			catch(err)

			if cmd != nil {
				wg := sync.WaitGroup{}

				stderr, err := cmd.StderrPipe()
				catch(err)
				buildErrBuf := &bytes.Buffer{}
				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := io.Copy(buildErrBuf, stderr)
					trace(err)
					err = stderr.Close()
					trace(err)
				}()

				err = cmd.Run()
				wg.Wait()
				if err != nil {
					exec.Build.Error = buildErrBuf.String()

					exec.Verdict = data.CompilationError

					exec.Status = 7
					err = exec.Put()
					catch(err)
					hub.Send([]interface{}{"SYNC", "executions", exec.Id})
					return
				}
			}

			exec.Status = 2
			err = exec.Put()
			catch(err)
			hub.Send([]interface{}{"SYNC", "executions", exec.Id})

			for i, test := range prob.Tests {
				inBlob, err := data.Blobs.Get(test.InputKey)
				catch(err)

				cmd = stack.Run(cell)
				catch(err)
				cmd.Limits.Cpu = time.Duration(prob.Limits.Cpu) * time.Second
				cmd.Limits.Memory = uint64(prob.Limits.Memory)

				wg := sync.WaitGroup{}

				stdin, err := cmd.StdinPipe()
				catch(err)
				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := io.Copy(stdin, inBlob)
					trace(err)
					err = inBlob.Close()
					trace(err)
					err = stdin.Close()
					trace(err)
				}()

				stdout, err := cmd.StdoutPipe()
				catch(err)
				outBuf := &bytes.Buffer{}
				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := io.Copy(outBuf, stdout)
					trace(err)
					err = stdout.Close()
					trace(err)
				}()

				err = cmd.Run()
				trace(err)
				wg.Wait()

				outKey := fmt.Sprintf("executions:%d:tests:%d:out", exec.Id, i)
				_, err = data.Blobs.Put(outKey, bytes.NewReader(outBuf.Bytes()))
				catch(err)

				diff := 0
				pts := 0
				if prob.Checker.Language == "" {
					ansBlob, err := data.Blobs.Get(test.AnswerKey)
					catch(err)
					for leftBuf, rightBuf := bufio.NewReader(ansBlob), bufio.NewReader(outBuf); ; {
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

					cmd = stack.Run(chkCell)
					catch(err)
					cmd.Limits.Cpu = 16 * time.Second
					cmd.Limits.Memory = 1 << 30

					wg := sync.WaitGroup{}

					stdin, err := cmd.StdinPipe()
					catch(err)
					wg.Add(1)
					go func() {
						defer wg.Done()

						_, err := io.Copy(stdin, outBuf)
						trace(err)
						err = stdin.Close()
						trace(err)
					}()

					stdout, err := cmd.StdoutPipe()
					catch(err)
					chkBuf := &bytes.Buffer{}
					wg.Add(1)
					go func() {
						defer wg.Done()

						_, err := io.Copy(chkBuf, stdout)
						trace(err)
						err = stdout.Close()
						trace(err)
					}()

					err = cmd.Run()
					catch(err)
					wg.Wait()

					n, _ := fmt.Fscanf(chkBuf, "%d %d", &diff, &pts)
					if n == 0 {
						diff = 1
					}
				}

				verdict := data.Verdict(0)
				cpu := float64(cmd.Usages.Cpu/1e6) / 1e3
				memory := int(cmd.Usages.Memory / (1 << 20))
				switch {
				case cpu > prob.Limits.Cpu:
					verdict = data.CpuLimitExceeded
					cpu = prob.Limits.Cpu

				case memory > prob.Limits.Memory:
					verdict = data.MemoryLimitExceeded
					memory = prob.Limits.Memory

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
						Cpu:    cpu,
						Memory: memory,
					},
					Points: pts,
				})

				err = exec.Put()
				catch(err)
				hub.Send([]interface{}{"SYNC", "executions", exec.Id})
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
			hub.Send([]interface{}{"SYNC", "executions", exec.Id})
		}()
	}
}

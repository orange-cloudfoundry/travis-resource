package messager_test

import (
	. "github.com/alphagov/travis-resource/messager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"bytes"
	"bufio"
	"github.com/mitchellh/colorstring"
	"fmt"
	"encoding/json"
	"errors"
)

var _ = Describe("Messager", func() {
	var (
		messager *ResourceMessager
		responseBuffer bytes.Buffer
		responseWritter *bufio.Writer
		logBuffer bytes.Buffer
		logWritter *bufio.Writer
	)
	BeforeEach(func() {
		responseBuffer.Reset()
		logBuffer.Reset()
		logWritter = bufio.NewWriter(&logBuffer)
		responseWritter = bufio.NewWriter(&responseBuffer)
		messager = NewMessager(logWritter, responseWritter)
		messager.SetExitOnFatal(false)
	})
	Describe("LogIt", func() {
		Context("When passing a simple text", func() {
			Context("without colors", func() {
				It("should output this text", func() {
					myText := "mytext"
					messager.LogIt(myText)
					logWritter.Flush()
					Expect(logBuffer.String()).To(Equal(myText))
				})
			})
			Context("with colors", func() {
				It("should output this text", func() {
					myText := "[red]mytext"
					messager.LogIt(myText)
					logWritter.Flush()
					Expect(logBuffer.String()).To(Equal(colorstring.Color(myText)))
				})
			})
		})
		Context("When passing a formatted text", func() {
			Context("without colors", func() {
				It("should output text formatted with good values", func() {
					myText := "mytext %s %d"
					number := 1
					messager.LogIt(myText, myText, number)
					logWritter.Flush()
					Expect(logBuffer.String()).To(Equal(fmt.Sprintf(myText, myText, number)))
				})
			})
			Context("with colors", func() {
				It("should output text formatted with good values", func() {
					myText := "[red]mytext %s %d"
					number := 1
					messager.LogIt(myText, myText, number)
					logWritter.Flush()
					Expect(logBuffer.String()).To(Equal(fmt.Sprintf(colorstring.Color(myText), myText, number)))
				})
			})
		})
		Context("When passing other thing than a string in first argument", func() {
			It("should panic", func() {
				actual := func() {
					messager.LogIt(1)
				}
				Expect(actual).Should(Panic())
			})
		})
	})
	Describe("LogItLn", func() {
		Context("When passing a simple text", func() {
			It("should output this text and appending a new line", func() {
				myText := "mytext"
				messager.LogItLn(myText)
				logWritter.Flush()
				Expect(logBuffer.String()).To(Equal(myText + "\n"))
			})
		})
		Context("When passing other thing than a string in first argument", func() {
			It("should panic", func() {
				actual := func() {
					messager.LogItLn(1)
				}
				Expect(actual).Should(Panic())
			})
		})
	})
	Describe("SendJsonResponse", func() {
		Context("when passing a struct with json template", func() {
			It("should output on response writer the formatted json", func() {
				type myStruct struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				}
				structToPass := &myStruct{"foo", "bar"}
				messager.SendJsonResponse(structToPass)
				responseWritter.Flush()
				var reponseJson myStruct
				err := json.Unmarshal(responseBuffer.Bytes(), &reponseJson)
				Expect(err).To(BeNil())
				Expect(reponseJson).To(BeEquivalentTo(reponseJson))
			})
		})
	})
	Describe("GetMessager", func() {
		Context("when asking multiple times a messager singleton", func() {
			It("should return always the same messager", func() {
				messager1 := GetMessager()
				messager2 := GetMessager()
				Expect(messager1).To(Equal(messager2))
			})
		})
	})
	Describe("Fatal", func() {
		Context("when writing an error message", func() {
			It("should write this message on response writer", func() {
				errorMessage := "error"
				messager.Fatal(errorMessage)
				responseWritter.Flush()
				Expect(responseBuffer.String()).To(Equal(errorMessage + "\n"))
			})
		})
	})
	Describe("FatalIf", func() {
		Context("when writing an error message and error is nil", func() {
			It("should not write this message on response writer", func() {
				errorMessage := "error"
				messager.FatalIf(errorMessage, nil)
				responseWritter.Flush()
				Expect(responseBuffer.String()).To(Equal(""))
			})
		})
		Context("when writing an error message and error is not nil", func() {
			It("should not write the message on response writer with error message", func() {
				errorMessage := "error"
				errorDetails := "it's an error"
				messager.FatalIf(errorMessage, errors.New(errorDetails))
				responseWritter.Flush()
				Expect(responseBuffer.String()).To(Equal(errorMessage + ": " + errorDetails + "\n"))
			})
		})
	})
})

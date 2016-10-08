package mock

import (
	"fmt"
	"strings"
)

// UserInterface represents a mock implementation of endive.UserInterface.
type UserInterface struct {
	UpdateValuesResult []string
}

// GetInput for mock UserInterface
func (u *UserInterface) GetInput() (string, error) {
	fmt.Println("mock UserInterface: GetInput")
	return "INPUT", nil
}

// YesOrNo for mock UserInterface
func (u *UserInterface) YesOrNo(a string) bool {
	fmt.Println("mock UserInterface: YesOrNo " + a)
	return true
}

// Choose for mock UserInterface
func (u *UserInterface) Choose(a, b, c, d string) (string, error) {
	fmt.Println("mock UserInterface: Choose " + a + ", " + b + ", " + c + ", " + d)
	return c, nil
}

// UpdateValues for mock UserInterface
func (u *UserInterface) UpdateValues(a, b string, c []string, isLong bool) ([]string, error) {
	fmt.Println("mock UserInterface: UpdateValues " + a + ", " + b + ", " + strings.Join(c, "|"))
	return u.UpdateValuesResult, nil
}

// Title for mock UserInterface
func (u *UserInterface) Title(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: Title " + a)
}

// SubTitle for mock UserInterface
func (u *UserInterface) SubTitle(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: SubTitle " + a)
}

// SubPart for mock UserInterface
func (u *UserInterface) SubPart(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: SubPart " + a)
}

// Choice for mock UserInterface
func (u *UserInterface) Choice(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: Choice " + a)
}

// Display for mock UserInterface
func (u *UserInterface) Display(a string) {
	fmt.Println("mock UserInterface: Display " + a)
}

// InitLogger for mock UserInterface
func (u *UserInterface) InitLogger(a string) error {
	fmt.Println("mock UserInterface: InitLogger " + a)
	return nil
}

// CloseLog for mock UserInterface
func (u *UserInterface) CloseLog() {
	fmt.Println("mock UserInterface: CloseLog ")
}

// Error for mock UserInterface
func (u *UserInterface) Error(a string) {
	fmt.Println("mock UserInterface: Error " + a)
}

// Warning for mock UserInterface
func (u *UserInterface) Warning(a string) {
	fmt.Println("mock UserInterface: Warning " + a)
}

// Info for mock UserInterface
func (u *UserInterface) Info(a string) {
	fmt.Println("mock UserInterface: Info " + a)
}

// Debug for mock UserInterface
func (u *UserInterface) Debug(a string) {
	fmt.Println("mock UserInterface: Debug " + a)
}

// Errorf for mock UserInterface
func (u *UserInterface) Errorf(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: Errorf " + a)
}

// Warningf for mock UserInterface
func (u *UserInterface) Warningf(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: Warningf " + a)
}

// Infof for mock UserInterface
func (u *UserInterface) Infof(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: Infof " + a)
}

// Debugf for mock UserInterface
func (u *UserInterface) Debugf(a string, b ...interface{}) {
	a = fmt.Sprintf(a, b...)
	fmt.Println("mock UserInterface: Debugf " + a)
}

// Edit for mock UserInterface
func (u *UserInterface) Edit(a string) (string, error) {
	fmt.Println("mock UserInterface: Edit " + a)
	return "edited", nil
}

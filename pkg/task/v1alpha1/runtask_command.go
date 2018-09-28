/*
Copyright 2018 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"
	"fmt"
	"time"

	. "github.com/openebs/maya/pkg/msg/v1alpha1"
)

// RunCommandAction determines the kind of action that gets executed by run task
// command
type RunCommandAction string

const (
	DeleteCommandAction RunCommandAction = "delete"
	CreateCommandAction RunCommandAction = "create"
	GetCommandAction    RunCommandAction = "get"
	ListCommandAction   RunCommandAction = "list"
	PatchCommandAction  RunCommandAction = "patch"
	UpdateCommandAction RunCommandAction = "update"
)

// RunCommandCategory represents the category of the runtask command
//
// NOTE:
//  A runtask command can belong to more than one categories
type RunCommandCategory string

const (
	JivaCommandCategory     RunCommandCategory = "jiva"
	CstorCommandCategory    RunCommandCategory = "cstor"
	VolumeCommandCategory   RunCommandCategory = "volume"
	PoolCommandCategory     RunCommandCategory = "pool"
	SnapshotCommandCategory RunCommandCategory = "snapshot"
)

// RunCommandCategoryList represents a list of RunCommandCategory
type RunCommandCategoryList []RunCommandCategory

func (l RunCommandCategoryList) String() string {
	return YamlString("runcommandcategories", l)
}

// Contains returns true if this list has the given category
func (l RunCommandCategoryList) Contains(given RunCommandCategory) (no bool) {
	if len(l) == 0 {
		return
	}
	for _, category := range l {
		if category == given {
			return !no
		}
	}
	return
}

// IsJivaVolume returns true if this list has both jiva and volume as its
// category items
func (l RunCommandCategoryList) IsJivaVolume() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(JivaCommandCategory) && l.Contains(VolumeCommandCategory) {
		return !no
	}
	return
}

// IsCstorVolume returns true if this list has both cstor and volume as its
// category items
func (l RunCommandCategoryList) IsCstorVolume() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(CstorCommandCategory) && l.Contains(VolumeCommandCategory) {
		return !no
	}
	return
}

// IsCstorSnapshot returns true if this list has both cstor and snapshot as its
// category items
func (l RunCommandCategoryList) IsCstorSnapshot() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(CstorCommandCategory) && l.Contains(SnapshotCommandCategory) {
		return !no
	}
	return
}

// IsValid returns true if category list is valid
func (l RunCommandCategoryList) IsValid() (no bool) {
	if len(l) == 0 {
		return
	}
	if l.Contains(JivaCommandCategory) && l.Contains(CstorCommandCategory) {
		// a volume can be either cstor or jiva based; not both
		return
	}
	return !no
}

// RunCommandData represents data provided to the runtask command before
// its execution i.e. input data
type RunCommandData interface{}

// RunCommandDataMap represents a map of input data required to execute
// runtask command
type RunCommandDataMap map[string]RunCommandData

func (m RunCommandDataMap) String() string {
	return YamlString("runcommanddatamap", m)
}

// RunCommandResult holds the result of executing runtask command
type RunCommandResult struct {
	Res    interface{} `json:"result"`          // result of runtask command execution
	Err    error       `json:"error"`           // root cause of issue; error if any during runtask command execution
	Extras AllMsgs     `json:"debug,omitempty"` // debug details i.e. errors, warnings, information, etc during execution
}

func NewRunCommandResult(result interface{}, extras AllMsgs) (r RunCommandResult) {
	return RunCommandResult{
		Res:    result,
		Err:    extras.Error(),
		Extras: extras,
	}
}

func (r RunCommandResult) String() string {
	return YamlString("runcommandresult", r)
}

func (r RunCommandResult) Error() error {
	return r.Err
}

func (r RunCommandResult) Result() interface{} {
	return r.Res
}

func (r RunCommandResult) Debug() AllMsgs {
	return r.Extras
}

var (
	NotSupportedCategoryError          = errors.New("not supported category: invalid runtask command")
	NotSupportedActionError            = errors.New("not supported action: invalid runtask command")
	InvalidCategoryError               = errors.New("invalid categories: invalid runtask command")
	CanNotRunDueToFailedConditionError = errors.New("can not execute runtask command: run condition failed")
)

// CommandRunner abstracts execution of runtask command
type CommandRunner interface {
	Run() (r RunCommandResult)
}

// RunCommand represent a runtask command
type RunCommand struct {
	ID          string                 // uniquely identifies a runtask command
	WillRun     bool                   // flags if this runtask command should get executed or not
	Action      RunCommandAction       // represents the runtask command's action
	Category    RunCommandCategoryList // classification of runtask command
	Data        RunCommandDataMap      // input data required to execute runtask command
	SelectPaths []string               // paths whose values will be retrieved after runtask command execution
	*Msgs                              // store and retrieve info, warns, errors, etc occurred during execution
}

// SelfInfo returns this instance of RunCommand as a string format
func (c *RunCommand) SelfInfo() string {
	var categories, data string
	for _, c := range c.Category {
		categories = categories + " " + string(c)
	}
	for n, d := range c.Data {
		data = data + fmt.Sprintf(" --%s=%s", n, d)
	}
	willrun := fmt.Sprintf(" --willrun=%t", c.WillRun)
	return fmt.Sprintf("%s%s%s%s", c.Action, categories, data, willrun)
}

// Command returns a new instance of RunCommand
func Command() *RunCommand {
	return &RunCommand{Msgs: &Msgs{}, WillRun: true}
}

// SetRun updates RunCommand instance with run flag
func (c *RunCommand) SetRun(willrun bool) (u *RunCommand) {
	c.WillRun = willrun
	return c
}

// IsRun flags if this runtask command will get executed or not
func (c *RunCommand) IsRun() bool {
	return c.WillRun
}

// AddError updates RunCommand instance with given error
func (c *RunCommand) AddError(err error) (u *RunCommand) {
	c.Msgs.AddError(err)
	return c
}

// CreateAction updates RunCommand instance with create action
func (c *RunCommand) CreateAction() (u *RunCommand) {
	c.Action = CreateCommandAction
	return c
}

// DeleteAction updates RunCommand instance with delete action
func (c *RunCommand) DeleteAction() (u *RunCommand) {
	c.Action = DeleteCommandAction
	return c
}

// GetAction updates RunCommand instance with get action
func (c *RunCommand) GetAction() (u *RunCommand) {
	c.Action = GetCommandAction
	return c
}

// ListAction updates RunCommand instance with list action
func (c *RunCommand) ListAction() (u *RunCommand) {
	c.Action = ListCommandAction
	return c
}

// UpdateAction updates RunCommand instance with update action
func (c *RunCommand) UpdateAction() (u *RunCommand) {
	c.Action = UpdateCommandAction
	return c
}

// PatchAction updates RunCommand instance with patch action
func (c *RunCommand) PatchAction() (u *RunCommand) {
	c.Action = PatchCommandAction
	return c
}

// WithCategory updates the given RunCommand instance with provided category
func WithCategory(given *RunCommand, category RunCommandCategory) (updated *RunCommand) {
	given.Category = append(given.Category, category)
	return given
}

// WithAction updates the given RunCommand instance with provided action
func WithAction(given *RunCommand, action RunCommandAction) (updated *RunCommand) {
	given.Action = action
	return given
}

// WithData updates the given RunCommand instance with provided input data
func WithData(given *RunCommand, name string, d RunCommandData) (updated *RunCommand) {
	if given.Data == nil {
		given.Data = map[string]RunCommandData{}
	}
	if d == nil {
		given.AddWarn(fmt.Sprintf("nil value provided for '%s': runtask command may fail", name))
	}
	given.Data[name] = d
	return given
}

// WithSelect updates the given RunCommand instance with provided select paths
func WithSelect(given *RunCommand, paths []string) (updated *RunCommand) {
	if len(paths) == 0 {
		return given
	}
	given.SelectPaths = append(given.SelectPaths, paths...)
	return given
}

func (c *RunCommand) String() string {
	return YamlString("runcommand", c)
}

func (c *RunCommand) Result(result interface{}) (r RunCommandResult) {
	return NewRunCommandResult(result, c.AllMsgs())
}

// instance fetches the specific runtask command implementation instance based
// on command categories
func (c *RunCommand) instance() (r CommandRunner) {
	if c.Category.IsJivaVolume() {
		r = &jivaVolumeCommand{cmd: c}
	} else if c.Category.IsCstorSnapshot() {
		r = &cstorSnapshotCommand{cmd: c}
	} else {
		r = &notSupportedCategoryCommand{cmd: c}
	}
	return
}

// preRun evaluates conditions and sets options prior to execution of runtask
// command
func (c *RunCommand) preRun() {
	if !c.Category.IsValid() {
		c.SetRun(false).AddError(InvalidCategoryError)
	}
	c.AddInfo(c.SelfInfo())
}

// Run finds the specific runtask command implementation and executes the same
func (c *RunCommand) Run() (r RunCommandResult) {
	c.preRun()
	if !c.IsRun() {
		return c.Result(nil)
	}
	return c.instance().Run()
}

// RunCommandMiddleware abstracts updating the given RunCommand instance
type RunCommandMiddleware func(given *RunCommand) (updated *RunCommand)

// JivaCategory updates RunCommand instance with jiva as the runtask command's
// category
func JivaCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, JivaCommandCategory)
	}
}

// CstorCategory updates RunCommand instance with cstor as the runtask command's
// category
func CstorCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, CstorCommandCategory)
	}
}

// VolumeCategory updates RunCommand instance with volume as the runtask
// command's category
func VolumeCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, VolumeCommandCategory)
	}
}

// SnapshotCategory updates RunCommand instance with snapshot as the runtask
// command's category
func SnapshotCategory() RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithCategory(given, SnapshotCommandCategory)
	}
}

// Select updates the RunCommand instance with paths whose values will be
// extracted after execution of runtask command
func Select(paths []string) RunCommandMiddleware {
	return func(given *RunCommand) (updated *RunCommand) {
		return WithSelect(given, paths)
	}
}

// RunCommandMiddlewareList represents a list of RunCommandMiddleware
type RunCommandMiddlewareList []RunCommandMiddleware

// Update updates the given RunCommand instance through all the middlewares
func (l RunCommandMiddlewareList) Update(given *RunCommand) (updated *RunCommand) {
	if len(l) == 0 || given == nil {
		return given
	}
	for _, middleware := range l {
		given = middleware(given)
	}
	return given
}

// notSupportedCategoryCommand is a CommandRunner implementation for
// un-supported runtask command category
type notSupportedCategoryCommand struct {
	cmd *RunCommand
}

func (c *notSupportedCategoryCommand) Run() (r RunCommandResult) {
	c.cmd.AddError(NotSupportedCategoryError)
	return NewRunCommandResult(nil, c.cmd.AllMsgs())
}

// notSupportedActionCommand is a CommandRunner implementation for
// un-supported runtask command action
type notSupportedActionCommand struct {
	cmd *RunCommand
}

func (c *notSupportedActionCommand) Run() (r RunCommandResult) {
	c.cmd.AddError(NotSupportedActionError)
	return NewRunCommandResult(nil, c.cmd.AllMsgs())
}

// ResultStoreFn abstracts storing a key value pair against the provided id
type ResultStoreFn func(id string, key string, value interface{})

// WillRunFn abstracts evaluation of condition to execute or skip a runtask
// command
type WillRunFn func() bool

// defaultCommandRunner manages execution of RunCommand instance based on
// provided hooks
//
// Hooks set in this runner are defined externally to this runner
//
// 1/ 'willrun' hook can be used to determine if the given runtask command
// should get executed or can be skipped
//
// 2/ 'store' hook can be used to save runtask command's execution response
type defaultCommandRunner struct {
	store   ResultStoreFn // mechanism to store runtask command's response
	willrun WillRunFn     // flag that determines if runtask command will execute or not
	id      string        // unique identification of the runtask command
	cmd     *RunCommand   // manages execution of runtask command
	*Msgs                 // store and retrieve info, warns, errors, etc occurred during execution
}

// DefaultCommandRunner returns a new instance of defaultCommandRunner
func DefaultCommandRunner(s ResultStoreFn, w WillRunFn) *defaultCommandRunner {
	return &defaultCommandRunner{store: s, willrun: w, Msgs: &Msgs{}}
}

// GetID returns unique id of this runner / runtask command
func (r *defaultCommandRunner) GetID() string {
	return r.id
}

// Result sets the result due to execution of runtask command
func (r *defaultCommandRunner) Result(id string, v interface{}) *defaultCommandRunner {
	r.store(id, "result", v)
	return r
}

// Debug sets the debug information due to execution of runtask command
func (r *defaultCommandRunner) Debug(id string, d AllMsgs) *defaultCommandRunner {
	r.store(id, "debug", d)
	return r
}

// Error sets the error due to execution of runtask command
func (r *defaultCommandRunner) Error(id string, e error) *defaultCommandRunner {
	r.store(id, "error", e)
	return r
}

// Stores sets various responses due to execution of runtask command
func (r *defaultCommandRunner) Stores(res RunCommandResult) {
	if len(r.id) == 0 {
		r.id = time.Now().Format("15040500000")
	}
	r.Result(r.id, res.Result()).Debug(r.id, res.Debug()).Error(r.id, res.Error())
}

// Command sets the runtask command to be executed via this runner
func (r *defaultCommandRunner) Command(id string, c *RunCommand) (u *defaultCommandRunner) {
	if len(id) == 0 {
		r.AddError(fmt.Errorf("missing runtask id: can not execute runtask command"))
		return r
	}
	r.id = id

	if c == nil {
		r.AddError(fmt.Errorf("nil runtask command: can not execute runtask command with id '%s'", id))
		return r
	}
	r.cmd = c
	return r
}

// preRun sets options conditionally prior to execution of runtask command
func (r *defaultCommandRunner) preRun() {
	if !r.willrun() && r.cmd != nil {
		r.cmd.SetRun(false).AddError(CanNotRunDueToFailedConditionError)
	}
}

// postRun executes the post activities associated after the execution of
// runtask command
func (r *defaultCommandRunner) postRun(res RunCommandResult) {
	r.Stores(res)
}

// Run executes the runtask command
func (r *defaultCommandRunner) Run() (res RunCommandResult) {
	r.preRun()
	if r.cmd == nil {
		res = NewRunCommandResult(nil, r.AllMsgs())
	} else {
		res = r.cmd.Run()
	}
	r.postRun(res)
	return
}

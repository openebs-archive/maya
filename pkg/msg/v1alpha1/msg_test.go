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
	"k8s.io/klog/v2"
	"testing"
)

func mockMsgFromType(mtype MsgType) *msg {
	return &msg{Mtype: mtype}
}

func mockMsgsFromType(mtypes []MsgType) Msgs {
	var msgs []*msg
	for _, mtype := range mtypes {
		msgs = append(msgs, mockMsgFromType(mtype))
	}
	return Msgs{Items: msgs}
}

func TestMsgPredicates(t *testing.T) {
	tests := map[string]struct {
		mtype     MsgType
		predicate msgPredicate
		expected  bool
	}{
		// tests against info message
		"101": {InfoMsg, IsInfo, true},
		"102": {InfoMsg, IsWarn, false},
		"103": {InfoMsg, IsErr, false},
		"104": {InfoMsg, IsNotInfo, false},
		"105": {InfoMsg, IsSkip, false},
		"106": {InfoMsg, IsNotErr, true},
		// tests against error message
		"201": {ErrMsg, IsInfo, false},
		"202": {ErrMsg, IsWarn, false},
		"203": {ErrMsg, IsErr, true},
		"204": {ErrMsg, IsNotInfo, true},
		"205": {ErrMsg, IsSkip, false},
		"206": {ErrMsg, IsNotErr, false},
		// tests against warn message
		"301": {WarnMsg, IsInfo, false},
		"302": {WarnMsg, IsWarn, true},
		"303": {WarnMsg, IsErr, false},
		"304": {WarnMsg, IsNotInfo, true},
		"305": {WarnMsg, IsSkip, false},
		"306": {WarnMsg, IsNotErr, true},
		// tests against skip message
		"401": {SkipMsg, IsInfo, false},
		"402": {SkipMsg, IsWarn, false},
		"403": {SkipMsg, IsErr, false},
		"404": {SkipMsg, IsNotInfo, true},
		"405": {SkipMsg, IsSkip, true},
		"406": {SkipMsg, IsNotErr, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m := mockMsgFromType(mock.mtype)
			if mock.expected != mock.predicate(m) {
				t.Fatalf("Test '%s' failed: expected '%t': actual '%t'", name, mock.expected, mock.predicate(m))
			}
		})
	}
}

func TestMsgsFilter(t *testing.T) {
	tests := map[string]struct {
		msgTypes  []MsgType
		predicate msgPredicate
		expected  int
	}{
		// filter cases against info messages
		"101": {[]MsgType{InfoMsg, WarnMsg, ErrMsg, SkipMsg}, IsInfo, 1},
		"102": {[]MsgType{InfoMsg, InfoMsg, ErrMsg, SkipMsg}, IsInfo, 2},
		"103": {[]MsgType{InfoMsg, InfoMsg, InfoMsg, SkipMsg}, IsInfo, 3},
		"104": {[]MsgType{InfoMsg, InfoMsg, InfoMsg, InfoMsg}, IsInfo, 4},
		"105": {[]MsgType{WarnMsg, ErrMsg, SkipMsg}, IsInfo, 0},
		"106": {[]MsgType{WarnMsg, ErrMsg}, IsInfo, 0},
		"107": {[]MsgType{WarnMsg}, IsInfo, 0},
		"108": {[]MsgType{}, IsInfo, 0},
		// filter cases against error messages
		"201": {[]MsgType{InfoMsg, WarnMsg, ErrMsg, SkipMsg}, IsErr, 1},
		"202": {[]MsgType{ErrMsg, ErrMsg, InfoMsg, SkipMsg}, IsErr, 2},
		"203": {[]MsgType{ErrMsg, ErrMsg, ErrMsg, SkipMsg}, IsErr, 3},
		"204": {[]MsgType{ErrMsg, ErrMsg, ErrMsg, ErrMsg}, IsErr, 4},
		"205": {[]MsgType{WarnMsg, InfoMsg, SkipMsg}, IsErr, 0},
		"206": {[]MsgType{WarnMsg, InfoMsg}, IsErr, 0},
		"207": {[]MsgType{WarnMsg}, IsErr, 0},
		"208": {[]MsgType{}, IsErr, 0},
		// filter cases against warn messages
		"301": {[]MsgType{InfoMsg, WarnMsg, ErrMsg, SkipMsg}, IsWarn, 1},
		"302": {[]MsgType{WarnMsg, WarnMsg, InfoMsg, SkipMsg}, IsWarn, 2},
		"303": {[]MsgType{WarnMsg, WarnMsg, WarnMsg, SkipMsg}, IsWarn, 3},
		"304": {[]MsgType{WarnMsg, WarnMsg, WarnMsg, WarnMsg}, IsWarn, 4},
		"305": {[]MsgType{ErrMsg, InfoMsg, SkipMsg}, IsWarn, 0},
		"306": {[]MsgType{ErrMsg, InfoMsg}, IsWarn, 0},
		"307": {[]MsgType{ErrMsg}, IsWarn, 0},
		"308": {[]MsgType{}, IsWarn, 0},
		// filter cases against skip messages
		"401": {[]MsgType{InfoMsg, WarnMsg, ErrMsg, SkipMsg}, IsSkip, 1},
		"402": {[]MsgType{SkipMsg, SkipMsg, InfoMsg, ErrMsg}, IsSkip, 2},
		"403": {[]MsgType{SkipMsg, SkipMsg, SkipMsg, ErrMsg}, IsSkip, 3},
		"404": {[]MsgType{SkipMsg, SkipMsg, SkipMsg, SkipMsg}, IsSkip, 4},
		"405": {[]MsgType{ErrMsg, InfoMsg, WarnMsg}, IsSkip, 0},
		"406": {[]MsgType{ErrMsg, InfoMsg}, IsSkip, 0},
		"407": {[]MsgType{ErrMsg}, IsSkip, 0},
		"408": {[]MsgType{}, IsSkip, 0},
		// filter cases against non error messages
		"501": {[]MsgType{InfoMsg, WarnMsg, ErrMsg, SkipMsg}, IsNotErr, 3},
		"502": {[]MsgType{SkipMsg, SkipMsg, InfoMsg, ErrMsg}, IsNotErr, 3},
		"503": {[]MsgType{SkipMsg, SkipMsg, SkipMsg, ErrMsg}, IsNotErr, 3},
		"504": {[]MsgType{SkipMsg, SkipMsg, SkipMsg, SkipMsg}, IsNotErr, 4},
		"505": {[]MsgType{ErrMsg, InfoMsg, WarnMsg}, IsNotErr, 2},
		"506": {[]MsgType{ErrMsg, InfoMsg}, IsNotErr, 1},
		"507": {[]MsgType{ErrMsg}, IsNotErr, 0},
		"508": {[]MsgType{}, IsNotErr, 0},
		// filter cases against non info messages
		"601": {[]MsgType{WarnMsg, WarnMsg, ErrMsg, SkipMsg}, IsNotInfo, 4},
		"602": {[]MsgType{SkipMsg, SkipMsg, InfoMsg, ErrMsg}, IsNotInfo, 3},
		"603": {[]MsgType{SkipMsg, SkipMsg, InfoMsg, InfoMsg}, IsNotInfo, 2},
		"604": {[]MsgType{SkipMsg, InfoMsg, InfoMsg, InfoMsg}, IsNotInfo, 1},
		"605": {[]MsgType{ErrMsg, InfoMsg, WarnMsg}, IsNotInfo, 2},
		"606": {[]MsgType{ErrMsg, InfoMsg}, IsNotInfo, 1},
		"607": {[]MsgType{InfoMsg}, IsNotInfo, 0},
		"608": {[]MsgType{}, IsNotInfo, 0},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m := mockMsgsFromType(mock.msgTypes)
			u := m.Filter(mock.predicate)
			if len(u.Items) != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%d': actual '%d'", name, mock.expected, len(u.Items))
			}
		})
	}
}

func TestMsgsLog(t *testing.T) {
	tests := map[string]struct {
		mtypes []MsgType
	}{
		"101": {[]MsgType{WarnMsg, ErrMsg, SkipMsg, InfoMsg}},
		"102": {[]MsgType{WarnMsg, ErrMsg, SkipMsg}},
		"103": {[]MsgType{WarnMsg, ErrMsg}},
		"104": {[]MsgType{WarnMsg}},
		"105": {[]MsgType{}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m := mockMsgsFromType(mock.mtypes)
			m.Log(klog.Infof)
		})
	}
}

func TestMsgsLogErrors(t *testing.T) {
	tests := map[string]struct {
		mtypes []MsgType
	}{
		"101": {[]MsgType{WarnMsg, ErrMsg, SkipMsg, InfoMsg}},
		"102": {[]MsgType{WarnMsg, ErrMsg, SkipMsg}},
		"103": {[]MsgType{WarnMsg, ErrMsg}},
		"104": {[]MsgType{WarnMsg}},
		"105": {[]MsgType{}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m := mockMsgsFromType(mock.mtypes)
			m.LogErrors(klog.Errorf)
		})
	}
}

func TestMsgsLogNonErrors(t *testing.T) {
	tests := map[string]struct {
		mtypes []MsgType
	}{
		"101": {[]MsgType{WarnMsg, ErrMsg, SkipMsg, InfoMsg}},
		"102": {[]MsgType{WarnMsg, ErrMsg, SkipMsg}},
		"103": {[]MsgType{WarnMsg, ErrMsg}},
		"104": {[]MsgType{WarnMsg}},
		"105": {[]MsgType{}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m := mockMsgsFromType(mock.mtypes)
			m.LogNonErrors(klog.Infof)
		})
	}
}

func TestMsgsLogNonInfos(t *testing.T) {
	tests := map[string]struct {
		mtypes []MsgType
	}{
		"101": {[]MsgType{WarnMsg, ErrMsg, SkipMsg, InfoMsg}},
		"102": {[]MsgType{WarnMsg, ErrMsg, SkipMsg}},
		"103": {[]MsgType{WarnMsg, ErrMsg}},
		"104": {[]MsgType{WarnMsg}},
		"105": {[]MsgType{}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			m := mockMsgsFromType(mock.mtypes)
			m.LogNonInfos(klog.Errorf)
		})
	}
}

func TestMsgsAddInfo(t *testing.T) {
	tests := map[string]struct {
		messages []string
		expected int
	}{
		"101": {[]string{"hi"}, 1},
		"102": {[]string{"hi", "there"}, 2},
		"103": {[]string{"hi", "there", "hello"}, 3},
		"104": {[]string{"hi", "there", "hello", "openebs"}, 4},
		"105": {[]string{"hi", "there", "hello", "openebs", "jiva"}, 5},
		"106": {[]string{"hi", "there", "hello", "openebs", "jiva", "cstor"}, 6},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ml := &Msgs{}
			for _, i := range mock.messages {
				ml.AddInfo(i)
			}
			if len(ml.Infos().Items) != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%d': actual '%d'", name, mock.expected, len(ml.Infos().Items))
			}
		})
	}
}

func TestMsgsAddWarn(t *testing.T) {
	tests := map[string]struct {
		messages []string
		expected int
	}{
		"101": {[]string{"hi"}, 1},
		"102": {[]string{"hi", "there"}, 2},
		"103": {[]string{"hi", "there", "hello"}, 3},
		"104": {[]string{"hi", "there", "hello", "openebs"}, 4},
		"105": {[]string{"hi", "there", "hello", "openebs", "jiva"}, 5},
		"106": {[]string{"hi", "there", "hello", "openebs", "jiva", "cstor"}, 6},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ml := &Msgs{}
			for _, i := range mock.messages {
				ml.AddWarn(i)
			}
			if len(ml.Warns().Items) != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%d': actual '%d'", name, mock.expected, len(ml.Warns().Items))
			}
		})
	}
}

func TestMsgsAddSkip(t *testing.T) {
	tests := map[string]struct {
		messages []string
		expected int
	}{
		"101": {[]string{"hi"}, 1},
		"102": {[]string{"hi", "there"}, 2},
		"103": {[]string{"hi", "there", "hello"}, 3},
		"104": {[]string{"hi", "there", "hello", "openebs"}, 4},
		"105": {[]string{"hi", "there", "hello", "openebs", "jiva"}, 5},
		"106": {[]string{"hi", "there", "hello", "openebs", "jiva", "cstor"}, 6},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ml := &Msgs{}
			for _, i := range mock.messages {
				ml.AddSkip(i)
			}
			if len(ml.Skips().Items) != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%d': actual '%d'", name, mock.expected, len(ml.Skips().Items))
			}
		})
	}
}

func TestMsgsAddError(t *testing.T) {
	tests := map[string]struct {
		messages []string
		expected int
	}{
		"101": {[]string{"hi"}, 1},
		"102": {[]string{"hi", "there"}, 2},
		"103": {[]string{"hi", "there", "hello"}, 3},
		"104": {[]string{"hi", "there", "hello", "openebs"}, 4},
		"105": {[]string{"hi", "there", "hello", "openebs", "jiva"}, 5},
		"106": {[]string{"hi", "there", "hello", "openebs", "jiva", "cstor"}, 6},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			ml := &Msgs{}
			for _, i := range mock.messages {
				ml.AddError(errors.New(i))
			}
			if len(ml.Errors().Items) != mock.expected {
				t.Fatalf("Test '%s' failed: expected '%d': actual '%d'", name, mock.expected, len(ml.Errors().Items))
			}
		})
	}
}

func TestMsgsMerge(t *testing.T) {
	err1 := fmt.Errorf("error1")
	err2 := fmt.Errorf("error2")

	tests := map[string]struct {
		existinfo     string
		existwarn     string
		existskip     string
		existerr      error
		newinfo       string
		newwarn       string
		newskip       string
		newerr        error
		expectedcount int
	}{
		"101": {"einfo", "ewarn", "eskip", err1, "ninfo", "nwarn", "nskip", err2, 8},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			e := &Msgs{}
			e.AddInfo(mock.existinfo)
			e.AddWarn(mock.existwarn)
			e.AddSkip(mock.existskip)
			e.AddError(mock.existerr)

			n := &Msgs{}
			n.AddInfo(mock.newinfo)
			n.AddWarn(mock.newwarn)
			n.AddSkip(mock.newskip)
			n.AddError(mock.newerr)

			e.Merge(n)
			if mock.expectedcount != len(e.Items) {
				t.Fatalf("Test '%s' failed: expected count %d: actual count %d", name, mock.expectedcount, len(e.Items))
			}
		})
	}
}

func TestAllMsgsToMsgs(t *testing.T) {
	tests := map[string]struct {
		info         string
		err          error
		warn         string
		skip         string
		expectedInfo int
		expectedErr  bool
		expectedWarn int
		expectedSkip int
	}{
		"101": {"i", errors.New("e"), "w", "s", 1, true, 1, 1},
		"102": {"", errors.New("e"), "w", "s", 0, true, 1, 1},
		"103": {"", nil, "w", "s", 0, false, 1, 1},
		"104": {"", nil, "", "s", 0, false, 0, 1},
		"105": {"", nil, "", "", 0, false, 0, 0},
		"106": {"i", nil, "", "", 1, false, 0, 0},
		"107": {"", errors.New("e"), "", "", 0, false, 0, 0},
		"108": {"", nil, "w", "", 0, false, 1, 0},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			o := &Msgs{}
			if len(mock.info) != 0 {
				o.AddInfo(mock.info)
			}
			if mock.err != nil {
				o.AddError(mock.err)
			}
			if len(mock.warn) != 0 {
				o.AddWarn(mock.warn)
			}
			if len(mock.skip) != 0 {
				o.AddSkip(mock.skip)
			}

			a := o.AllMsgs()
			n := a.ToMsgs()

			if n == nil {
				t.Fatalf("Test '%s' failed: expected not nil msgs: actual nil msgs", name)
			}
			if len(n.Infos().Items) != mock.expectedInfo {
				t.Fatalf("Test '%s' failed: expected infos %d: actual infos %d", name, mock.expectedInfo, len(n.Infos().Items))
			}
			if mock.expectedErr && len(n.Errors().Items) != 1 {
				t.Fatalf("Test '%s' failed: expected 1 error: actual errors %d", name, len(n.Errors().Items))
			}
			if len(n.Skips().Items) != mock.expectedSkip {
				t.Fatalf("Test '%s' failed: expected skips %d: actual skips %d", name, mock.expectedSkip, len(n.Skips().Items))
			}
			if len(n.Warns().Items) != mock.expectedWarn {
				t.Fatalf("Test '%s' failed: expected warns %d: actual warns %d", name, mock.expectedWarn, len(n.Warns().Items))
			}
		})
	}
}

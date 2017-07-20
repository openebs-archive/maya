package command_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openebs/maya/command"
	"strconv"
)

var _ = Describe("VsmStats", func() {

	c := VsmStatsCommand{
		Address:     "AvailableAtRunTime",
		Host:        "false",
		Length:      0,
		Replica_ips: "0",
		Json:        "json",
	}

	d := VsmStatsCommand{
		Address:     "AvailableAtRunTime",
		Host:        "false",
		Length:      0,
		Replica_ips: "0",
		Json:        "",
	}

	a := Annotations{
		Iqn:          "iqn.2016-09.com.openebs.jiva:vol",
		TargetPortal: "10.99.73.74:3260",
		VolSize:      "1G",
	}

	arg := []string{"vol"}

	starr := make([]string, 0)
	starr = append(starr, fmt.Sprintf("%-15s %-10s %s", "10.44.0.1", "Online", "0"))
	starr = append(starr, fmt.Sprintf("%-15s %-10s %s", "10.36.0.1", "Online", "0"))

	stat1 := VolumeStats{
		ReadIOPS:             strconv.Itoa(0),
		WriteIOPS:            strconv.Itoa(0),
		TotalReadTime:        strconv.Itoa(0),
		TotalReadBlockCount:  strconv.Itoa(0),
		TotalWriteTime:       strconv.Itoa(0),
		TotalWriteBlockCount: strconv.Itoa(0),
		SectorSize:           strconv.Itoa(0),
		UsedBlocks:           strconv.Itoa(0),
		UsedLogicalBlocks:    strconv.Itoa(0),
	}

	stat2 := VolumeStats{
		ReadIOPS:             strconv.Itoa(1024),
		WriteIOPS:            strconv.Itoa(1024),
		TotalReadTime:        strconv.Itoa(1024),
		TotalReadBlockCount:  strconv.Itoa(1024),
		TotalWriteTime:       strconv.Itoa(1024),
		TotalWriteBlockCount: strconv.Itoa(1024),
		SectorSize:           strconv.Itoa(4096),
		UsedBlocks:           strconv.Itoa(10),
		UsedLogicalBlocks:    strconv.Itoa(10),
	}

	Context("Command line argument", func() {

		It("has some default values with json", func() {
			Expect(c.Address).To(Equal("AvailableAtRunTime"))
			Expect(c.Host).To(Equal("false"))
			Expect(c.Length).To(Equal(0))
			Expect(c.Replica_ips).To(Equal("0"))
			Expect(c.Json).To(Equal("json"))
		})
	})

	Context("Command line argument", func() {

		It("has some default values with json", func() {
			Expect(d.Address).To(Equal("AvailableAtRunTime"))
			Expect(d.Host).To(Equal("false"))
			Expect(d.Length).To(Equal(0))
			Expect(d.Replica_ips).To(Equal("0"))
			Expect(d.Json).To(Equal(""))
		})
	})

	Context("Annotations", func() {

		It("has some default values", func() {
			Expect(a.Iqn).To(Equal("iqn.2016-09.com.openebs.jiva:vol"))
			Expect(a.TargetPortal).To(Equal("10.99.73.74:3260"))
			Expect(a.VolSize).To(Equal("1G"))
		})
	})

	Context("Status Array", func() {

		It("has some default values", func() {
			Expect(starr[0]).To(Equal("10.44.0.1       Online     0"))
			Expect(starr[1]).To(Equal("10.36.0.1       Online     0"))
		})
	})

	Context("Volume Stats 1", func() {
		It("has some default values", func() {
			Expect(stat1.ReadIOPS).To(Equal("0"))
			Expect(stat1.WriteIOPS).To(Equal("0"))
			Expect(stat1.TotalReadTime).To(Equal("0"))
			Expect(stat1.TotalReadBlockCount).To(Equal("0"))
			Expect(stat1.TotalWriteTime).To(Equal("0"))
			Expect(stat1.TotalWriteBlockCount).To(Equal("0"))
			Expect(stat1.SectorSize).To(Equal("0"))
			Expect(stat1.UsedBlocks).To(Equal("0"))
			Expect(stat1.UsedLogicalBlocks).To(Equal("0"))
		})
	})

	Context("Volume Stats 2", func() {
		It("has some default values", func() {
			Expect(stat2.ReadIOPS).To(Equal("1024"))
			Expect(stat2.WriteIOPS).To(Equal("1024"))
			Expect(stat2.TotalReadTime).To(Equal("1024"))
			Expect(stat2.TotalReadBlockCount).To(Equal("1024"))
			Expect(stat2.TotalWriteTime).To(Equal("1024"))
			Expect(stat2.TotalWriteBlockCount).To(Equal("1024"))
			Expect(stat2.SectorSize).To(Equal("4096"))
			Expect(stat2.UsedBlocks).To(Equal("10"))
			Expect(stat2.UsedLogicalBlocks).To(Equal("10"))

		})
	})

	Context("Testing the StatusVolume with json", func() {
		It("Passing the default values", func() {
			Expect(StatsOutput(&c, &a, arg, starr, stat1, stat2)).Should(BeZero())

		})

	})

	Context("Testing the StatusVolume with default", func() {
		It("Passing the default values", func() {
			Expect(StatsOutput(&d, &a, arg, starr, stat1, stat2)).Should(BeZero())

		})

	})

})

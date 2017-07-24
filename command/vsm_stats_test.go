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
		Address:     "Address",
		Host:        "false",
		Length:      0,
		Replica_ips: "0",
		Json:        "json",
	}

	d := VsmStatsCommand{
		Address:     "Address",
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
	replica := []string{"10.44.0.1", "10.36.0.1"}
	status := []string{"Online", "Online"}
	dataUpdateIndex := []string{"0", "0"}

	starr := make([]string, 0)
	for i := range replica {

		starr = append(starr, fmt.Sprintf("%s", replica[i]))
		starr = append(starr, fmt.Sprintf("%s", status[i]))
		starr = append(starr, fmt.Sprintf("%s", dataUpdateIndex[i]))
	}

	const random1 int = 0
	const random2 int = 1024
	const random3 int = 4096
	const random int = 10
	stat1 := VolumeStats{
		ReadIOPS:             strconv.Itoa(random1),
		WriteIOPS:            strconv.Itoa(random1),
		TotalReadTime:        strconv.Itoa(random1),
		TotalReadBlockCount:  strconv.Itoa(random1),
		TotalWriteTime:       strconv.Itoa(random1),
		TotalWriteBlockCount: strconv.Itoa(random1),
		SectorSize:           strconv.Itoa(random1),
		UsedBlocks:           strconv.Itoa(random1),
		UsedLogicalBlocks:    strconv.Itoa(random1),
	}

	stat2 := VolumeStats{
		ReadIOPS:             strconv.Itoa(random2),
		WriteIOPS:            strconv.Itoa(random2),
		TotalReadTime:        strconv.Itoa(random2),
		TotalReadBlockCount:  strconv.Itoa(random2),
		TotalWriteTime:       strconv.Itoa(random2),
		TotalWriteBlockCount: strconv.Itoa(random2),
		SectorSize:           strconv.Itoa(random3),
		UsedBlocks:           strconv.Itoa(random),
		UsedLogicalBlocks:    strconv.Itoa(random),
	}

	Context("Command line argument", func() {

		It("has some default values with json", func() {
			Expect(c.Address).To(Equal("Address"))
			Expect(c.Host).To(Equal("false"))
			Expect(c.Length).To(Equal(0))
			Expect(c.Replica_ips).To(Equal("0"))
			Expect(c.Json).To(Equal("json"))
		})
	})

	Context("Command line argument", func() {

		It("has some default values with json", func() {
			Expect(d.Address).To(Equal("Address"))
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
			Expect(starr[0]).To(Equal("10.44.0.1"))
			Expect(starr[3]).To(Equal("10.36.0.1"))
			Expect(starr[1]).To(Equal("Online"))
			Expect(starr[4]).To(Equal("Online"))
			Expect(starr[2]).To(Equal("0"))
			Expect(starr[5]).To(Equal("0"))
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

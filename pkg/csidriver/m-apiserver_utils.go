package driver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func provisionVolume(req *csi.CreateVolumeRequest) (*v1alpha1.CASVolume, error) {
	casVolume := v1alpha1.CASVolume{}

	capacity := int64(req.GetCapacityRange().GetRequiredBytes())
	if capacity >= maxStorageCapacity {
		return nil, status.Errorf(codes.OutOfRange,
			"Requested capacity %d exceeds maximum allowed %d",
			capacity, maxStorageCapacity)
	}
	casVolume.Spec.Capacity = strconv.FormatInt(capacity, 10)

	parameters := req.GetParameters()
	storageclass := parameters["storageclass"]
	namespace := parameters["namespace"]
	// creating a map b/c have to initialize the map using the make function
	// before adding any elements to avoid nil map assignment error
	mapLabels := make(map[string]string)

	if storageclass == "" {
		logrus.Errorf("Volume has no storage class specified")
	} else {
		mapLabels[string(v1alpha1.StorageClassKey)] = storageclass
		casVolume.Labels = mapLabels
	}

	casVolume.Labels[string(v1alpha1.NamespaceKey)] = namespace
	casVolume.Namespace = namespace
	casVolume.Labels[string(v1alpha1.PersistentVolumeClaimKey)] =
		parameters["persistentvolumeclaim"]
	casVolume.Name = req.GetName()

	// Check if volume already exists
	// if present then return the read values
	// if unexpected error then return the error
	// if absent then create volume
	logrus.Infof("Checking if volume %q already exists", casVolume.Name)
	err := ReadVolume(req.GetName(), namespace, storageclass, &casVolume)
	if err == nil {
		logrus.Infof("Volume %v already present", req.GetName())
	} else if err.Error() != http.StatusText(404) {
		// any error other than 404 is unexpected error
		logrus.Errorf(
			"Unexpected error occurred while trying to read the volume: %s", err)
		return nil, err
	} else if err.Error() == http.StatusText(404) {
		// Create the volume and read it
		logrus.Infof("Volume %q does not exist,attempting to create volume",
			req.GetName)
		err = CreateVolume(casVolume)
		if err != nil {
			logrus.Errorf("Failed to create volume:  %+v, error: %s",
				req.GetName, err.Error())
			return nil, err
		}
		err = ReadVolume(req.GetName(), namespace, storageclass, &casVolume)
		if err != nil {
			logrus.Errorf("Failed to read volume: %v", err)
			return nil, err
		}
		logrus.Infof("VolumeInfo: created volume metadata : %#v", casVolume)
	}
	return &casVolume, nil
}

// CreateVolume to create the CAS volume through a API call to m-apiserver
func CreateVolume(vol v1alpha1.CASVolume) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		return err
	}
	url := addr + "/latest/volumes/"

	//Marshal serializes the value provided into a json document
	jsonValue, _ := json.Marshal(vol)

	logrus.Infof("CAS Volume Spec Created:\n%v\n", string(jsonValue))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))

	req.Header.Add("Content-Type", "application/json")

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		logrus.Errorf("Error when connecting maya-apiserver %v", err)
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Unable to read response from maya-apiserver %v", err)
		return err
	}
	code := resp.StatusCode
	if code != http.StatusOK {
		logrus.Errorf("%s: failed to create volume '%s': response: %+v",
			http.StatusText(code), vol.Name, string(data))
		return fmt.Errorf("%s: failed to create volume '%s': response: %+v",
			http.StatusText(code), vol.Name, string(data))
	}

	logrus.Infof("volume Successfully Created:\n%+v", string(data))
	return nil
}

// ReadVolume to get the info of CAS volume through a API call to m-apiserver
func ReadVolume(vname, namespace, storageclass string, obj interface{}) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		return err
	}
	url := addr + "/latest/volumes/" + vname

	logrus.Infof("[DEBUG] Get details for Volume :%v", string(vname))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("namespace", namespace)
	// passing storageclass info as a request header which will extracted by the
	// Maya-apiserver to get the CAS template name
	req.Header.Set(string(v1alpha1.StorageClassHeaderKey), storageclass)

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		logrus.Errorf("Error when connecting to maya-apiserver %v", err)
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		logrus.Errorf("HTTP Status error from maya-apiserver: %v\n",
			http.StatusText(code))
		return errors.New(http.StatusText(code))
	}
	logrus.Info("volume Details Successfully Retrieved")
	return json.NewDecoder(resp.Body).Decode(obj)
}

// DeleteVolume to get delete CAS volume through a API call to m-apiserver
func DeleteVolume(vname string, namespace string) error {

	addr := os.Getenv("MAPI_ADDR")
	if addr == "" {
		err := errors.New("MAPI_ADDR environment variable not set")
		return err
	}
	url := addr + "/latest/volumes/" + vname
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("namespace", namespace)

	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		logrus.Errorf("Error when connecting to maya-apiserver  %v", err)
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		return fmt.Errorf("failed to delete volume %s:%s",
			vname, http.StatusText(code))
	}
	logrus.Info("volume Deleted Successfully initiated")
	return nil
}

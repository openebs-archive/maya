#include <libudev.h>
#include <stdio.h>
#include <errno.h>

int main(int argc, char **argv)
{
	char *device;
	struct udev *udev;
	struct udev_device *dev = NULL;
	int initialized = -1;
	int rc = -1;

	if ((udev = udev_new()) == NULL) {
	        fprintf(stderr, "failed to get udev device\n");
	        return (-1);
	}

	if (argc > 1) {
	        device = argv[1];
	} else {
	        fprintf(stderr, "device is not provided\n");
	        return (-1);
	}

	dev = udev_device_new_from_subsystem_sysname(udev, "block", device);

	/* https://github.com/jcnelson/vdev/blob/ceb7a6c4f44dec542dc1c3c3d5abd27dec7f3e0e/libudev-compat/libudev-device.c#L1772
	 * From above code comments udev_device_get_is_initialized
	 * Returns: 1 if the device is set up. 0 otherwise.
	 */
	if ((dev != NULL) && (rc = udev_device_get_is_initialized(dev)) == 1) {
		printf("device = %s is initialized by udev\n", device);
		initialized = 0;
	} else {
	        printf("device = %s is not initialized by udev errno: %d\n", device, errno);
	}

	udev_device_unref(dev);
	udev_unref(udev);
	return initialized;
}


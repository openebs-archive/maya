#include <libudev.h>
#include <stdio.h>

int main(int argc, char **argv)
{
        char *device;
        struct udev *udev;
        struct udev_device *dev = NULL;

        if ((udev = udev_new()) == NULL) {
                printf("failed to get udev device\n");
                return (-1);
        }

        if (argc > 1) {
                device = argv[1];
        } else {
                fprintf(stderr, "device is not provided\n");
                return (-1);
        }

        dev = udev_device_new_from_subsystem_sysname(udev, "block", device);

        if ((dev != NULL) && udev_device_get_is_initialized(dev)) {
                printf("device = %s is initialized by udev\n", device);
        } else {
                printf("device = %s is not initialized by udev\n", device);
        }

        udev_device_unref(dev);
        udev_unref(udev);
        return (0);
}


package drivers

import (
	"errors"

	"github.com/canonical/lxd/shared/ioprogress"
)

// createVolumeFromImage creates a new volume from an image. If imgVol is nil (no cached
// image volume is available), it falls back to unpacking the image directly into a new
// volume via the filler. Otherwise it clones from imgVol, falling back to a direct unpack
// if imgVol's config is incompatible with vol or the image volume cannot be shrunk to
// the requested size.
func createVolumeFromImage(vol Volume, imgVol *Volume, filler *VolumeFiller, progressReporter ioprogress.ProgressReporter) error {
	// vol is passed by value so it is never nil, but its driver field may be unset.
	if vol.driver == nil {
		return errors.New("Volume has no associated driver")
	}

	// Image-only mode: caller asks the driver to materialise the image volume itself
	// (vol is the image volume, no source to clone from). Skip if already on disk.
	if imgVol == nil && vol.volType == VolumeTypeImage {
		exists, err := vol.driver.HasVolume(vol)
		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		return vol.driver.CreateVolume(vol, filler, progressReporter)
	}

	if imgVol == nil {
		return vol.driver.CreateVolume(vol, filler, progressReporter)
	}

	// imgVol must be compatible with vol's effective config. For variant-aware drivers
	// the caller has already selected the right variant; for single-variant drivers
	// imgVol is the pool-default cached image, so this also catches per-instance
	// config that diverges from pool defaults.
	if !vol.driver.ImageVolumeConfigMatch(*imgVol, vol) {
		return vol.driver.CreateVolume(vol, filler, progressReporter)
	}

	// Derive the volume size to use for the new volume when copying from the image volume.
	// Where possible (if the image volume has a volatile.rootfs.size property), it checks
	// that the image volume isn't larger than the volume's "size" and the pool's
	// "volume.size" setting.
	newVolSize, err := vol.ConfigSizeFromSource(*imgVol)
	if err != nil {
		return err
	}

	vol.SetConfigSize(newVolSize)

	// Clone the new volume from the cached image volume.
	err = vol.driver.CreateVolumeFromCopy(NewVolumeCopy(vol), NewVolumeCopy(*imgVol), false, progressReporter)
	if err != nil {
		if !errors.Is(err, ErrCannotBeShrunk) {
			return err
		}

		// The cached image volume is larger than the requested new volume size and cannot
		// be shrunk. Fall back to unpacking the image directly into a new volume. This is
		// slower but allows creating volumes smaller than the pool's volume settings.
		return vol.driver.CreateVolume(vol, filler, progressReporter)
	}

	return nil
}

// FUSE main Fs

// +build linux darwin freebsd

package mount

import (
	"syscall"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/ncw/rclone/fs"
	"github.com/ncw/rclone/vfs"
	"github.com/ncw/rclone/vfs/vfsflags"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// FS represents the top level filing system
type FS struct {
	*vfs.VFS
	f fs.Fs
}

// Check interface satistfied
var _ fusefs.FS = (*FS)(nil)

// NewFS makes a new FS
func NewFS(f fs.Fs) *FS {
	fsys := &FS{
		VFS: vfs.New(f, &vfsflags.Opt),
		f:   f,
	}
	return fsys
}

// Root returns the root node
func (f *FS) Root() (node fusefs.Node, err error) {
	defer fs.Trace("", "")("node=%+v, err=%v", &node, &err)
	root, err := f.VFS.Root()
	if err != nil {
		return nil, translateError(err)
	}
	return &Dir{root}, nil
}

// Check interface satsified
var _ fusefs.FSStatfser = (*FS)(nil)

// Statfs is called to obtain file system metadata.
// It should write that data to resp.
func (f *FS) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) (err error) {
	defer fs.Trace("", "")("stat=%+v, err=%v", resp, &err)
	const blockSize = 4096
	const fsBlocks = (1 << 50) / blockSize
	resp.Blocks = fsBlocks  // Total data blocks in file system.
	resp.Bfree = fsBlocks   // Free blocks in file system.
	resp.Bavail = fsBlocks  // Free blocks in file system if you're not root.
	resp.Files = 1E9        // Total files in file system.
	resp.Ffree = 1E9        // Free files in file system.
	resp.Bsize = blockSize  // Block size
	resp.Namelen = 255      // Maximum file name length?
	resp.Frsize = blockSize // Fragment size, smallest addressable data size in the file system.
	return nil
}

// Translate errors from mountlib
func translateError(err error) error {
	if err == nil {
		return nil
	}
	switch errors.Cause(err) {
	case vfs.OK:
		return nil
	case vfs.ENOENT:
		return fuse.ENOENT
	case vfs.EEXIST:
		return fuse.EEXIST
	case vfs.EPERM:
		return fuse.EPERM
	case vfs.ECLOSED:
		return fuse.Errno(syscall.EBADF)
	case vfs.ENOTEMPTY:
		return fuse.Errno(syscall.ENOTEMPTY)
	case vfs.ESPIPE:
		return fuse.Errno(syscall.ESPIPE)
	case vfs.EBADF:
		return fuse.Errno(syscall.EBADF)
	case vfs.EROFS:
		return fuse.Errno(syscall.EROFS)
	case vfs.ENOSYS:
		return fuse.ENOSYS
	}
	return err
}

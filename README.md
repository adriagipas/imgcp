# imgcp

**imgcp** is a command line tool that allows to copy files between old
  disk images (floppies, hard drives, archive files, etc). Currently
  supported formats are:

 - 3DS file formats (3DS/CCI, NCCH/CXI) (*read only*) 
 - CD images (CUE/BIN, MDS/MDF) (*read only*)
 - FAT12
 - FAT16
 - Interchange File Format (IFF) files (*read only*)
 - ISO 9660 (*read only*)

Apart from copying files, **imgcp** also implements other useful operations:

 - **cat**: Similar to the UNIX *cat* command, it can be used to print
     on the standard output the concatenation of several files inside
     disk images.
 - **ls**: Similar to the UNIX *ls* command, it can be used to explore
     the content of an image.
 - **mkdir**: To create empty directories.
 - **remove**: To remove files and directories.
 - **show**: The default operation. It shows basic information of the
     input images.
     
## Installing imgcp

First get the repository:
```
git clone https://github.com/adriagipas/imgcp.git
cd imgcp
```

Then install the software using the *go* command:
```
go build
go install
```
This will install **imgcp** to your standard *Go install
directory*. If you want to change the standard directory please refer
to the official [compile-install
tutorial](https://go.dev/doc/tutorial/compile-install)

## Examples

Print basic version and usage information:
```
imgcp
```

Print basic information of a hard drive image (*hdd.img*):
```
imgcp hdd.img
```

List the contents of the first partition of a hard drive image (*hdd.img*):
```
imgcp hdd.img ls /0
```

List the contents of the *DOS* folder in the first partition of a hard
drive image (*hdd.img*):
```
imgcp hdd.img ls /0/DOS
```

Concatenate the content of *AUTOEXEC.BAT* and *CONFIG.SYS* files
from first partition of *hdd.img*:
```
imgcp hdd.img cat /0/autoexec.bat /0/config.sys
```

In the previous hard drive image, copy *AUTOEXEC.BAT* as "AUTOCOP.BAT"
to a new empty folder *FOO*:
```
imgcp hdd.img mkdir /0/foo
imgcp hdd.img cp /0/autoexec.bat /0/foo/autocop.bat
```

In previous folder *FOO*, copy *CONFIG.SYS* and *DOS* folder:
```
imgcp hdd.img cp /0/config.sys /0/DOS /0/foo/
```

Copy the content of an old floppy image (*floppy.img*) into folder
*DISK* in the first partition of *hdd.img*:
```
imgcp hdd.img mkdir /0/DISK
imgcp A=hdd.img B=floppy.img cp B=/ A=/0/DISK
```

Remove *FOO* and *DISK* folders from previous examples:
```
imgcp hdd.img rm /0/disk /0/foo
```

Copy the content of an old floppy image (*floppy.image*) into a local
folder */tmp/disk*:
```
mkdir -p /tmp/disk
imgcp A=floppy.img B=/ cp A=/ B=/tmp/disk/
```

Copy previous folder */tmp/disk* into the first partition of
*hdd.img*:
```
imgcp A=hdd.img B=/ cp B=/tmp/disk A=/0/
```

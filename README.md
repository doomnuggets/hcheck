# hcheck

*One of my first Go projects, terrible code.*

hcheck takes a sha256sum file input and checks if the hashes either matches,
mismatches, are new or have been removed.


Given the following directory structure...

```
tst
├── file1.bin
└── sub
    └── directory
        └── file2.bin

2 directories, 2 files
```

We first create a sha256sum of the directory tree:

```bash
find tst/ -type f -exec sha256sum {} \; > hashes.txt

cat hashes.txt
0fb2eb1d47a1978e2e019e795bca83b758847d590fdef757f749dd44358cc4ef  tst/file1.bin
33b999f808fda86a6bb9cb583a97c66775a6f9bd3602c4cceb27b235d697c7e3  tst/sub/directory/file2.bin
```

We then check if the hashes we just created match:

```
./hcheck --check-dir tst/ --hash-file hashes.txt
0fb2eb1d47a1978e2e019e795bca83b758847d590fdef757f749dd44358cc4ef  tst/file1.bin: OK
33b999f808fda86a6bb9cb583a97c66775a6f9bd3602c4cceb27b235d697c7e3  tst/sub/directory/file2.bin: OK
```

As we can see all hashes matched up (status `OK`).
Let's try and add a new file and modify `file1.bin`

```
echo new_file > tst/sub/new_file.bin
echo change >> tst/file1.bin
```

Now the output of hcheck shows:

```
./hcheck --check-dir tst/ --hash-file hashes.txt

e9c3d6e78375b7350ae37cac2ce6040b2bbbfee92440e9cfb7b461643e2a170e  tst/file1.bin: MISMATCH
33b999f808fda86a6bb9cb583a97c66775a6f9bd3602c4cceb27b235d697c7e3  tst/sub/directory/file2.bin: OK
294e1ef3296ec3b9e19a4acd0ecd3344aff767e7529eec0e2295bb7f69ca13f8  tst/sub/new_file.bin: NEW
```

Upon removing a file recorded in the hash file (hashes.txt) we see the following:

```
rm tst/sub/directory/file2.bin

./hcheck --check-dir tst/ --hash-file hashes.txt
e9c3d6e78375b7350ae37cac2ce6040b2bbbfee92440e9cfb7b461643e2a170e  tst/file1.bin: MISMATCH
294e1ef3296ec3b9e19a4acd0ecd3344aff767e7529eec0e2295bb7f69ca13f8  tst/sub/new_file.bin: NEW
33b999f808fda86a6bb9cb583a97c66775a6f9bd3602c4cceb27b235d697c7e3  tst/sub/directory/file2.bin: REMOVED
```


# Why

With the information about matching, changed, new and removed files we can keep track of a specific directory.
I use this in a honey pot environment I'm working on, to detect changes to the filesystem and trigger scripts.


# License

```

            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
                    Version 2, December 2004

 Copyright (C) 2018 doomnuggets

 Everyone is permitted to copy and distribute verbatim or modified
 copies of this license document, and changing it is allowed as long
 as the name is changed.

            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
   TERMS AND CONDITIONS FOR COPYING, DISTRIBUTION AND MODIFICATION

  0. You just DO WHAT THE FUCK YOU WANT TO.

```


# Fork me

Use the fork luke.

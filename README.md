go-arff2root
============

``go-arff2root`` is a simple command to convert an `ARFF` file into a `ROOT` file, converting the `ARFF` relation into a `ROOT` ``TTree``.

## Installation

```sh
$ go get github.com/sbinet/go-arff2root
```

## Example

```sh
$ go-arff2root data.arff data.root
$ go-arff2root -i data.arff -o data.root
$ go-arff2root data.arff.gz data.root
```


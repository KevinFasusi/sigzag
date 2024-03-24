package crawler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/KevinFasusi/hometree"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type FileType int

const (
	Manifest FileType = iota
	MerkleTree
)

func (f FileType) Strings() string {
	return [...]string{
		"manifest",
		"merkletree",
	}[f]
}

type Config struct {
	Ext  string
	Size int64
	//HashType hashType
}

type Crawler interface {
	Crawl(root string, out io.Writer) error
}

type DirectoryCrawler struct {
	Dir         string           `json:"dir"`
	Regex       []*regexp.Regexp `json:"regex"`
	Conf        Config           `json:"conf"`
	FileDigests [][]byte         `json:"file_digests"`
	Signatures  []*Sig           `json:"signatures"`
}

type Sig struct {
	Asset  string `json:"asset"`
	Digest string `json:"sha256"`
}

func (d *DirectoryCrawler) Crawl() error {

	return filepath.WalkDir(d.Dir, d.signatureWalk)
}

func (d *DirectoryCrawler) signatureWalk(path string, info fs.DirEntry, err error) error {
	if !info.IsDir() {
		go func() {
			signature := d.FileSignature(path)
			d.FileDigests = append(d.FileDigests, signature)
			s := Sig{Asset: path, Digest: fmt.Sprintf("%x", signature)}
			d.Signatures = append(d.Signatures, &s)
		}()
	}
	return nil
}

func (d *DirectoryCrawler) FileSignature(path string) []byte {
	var sum []byte
	file, err := os.Open(path)

	if err != nil {
		fmt.Printf("ERR\n")
	}

	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal()
		}
	}(file)

	buf := make([]byte, 8192)
	fileSignature := sha256.New()

	for b := 0; err == nil; {
		b, err = file.Read(buf)
		if err == nil {
			_, err = fileSignature.Write(buf[:b])
		}
	}
	sum = fileSignature.Sum(nil)
	return sum
}

func (d *DirectoryCrawler) Write(fileType FileType) error {
	s := sha256.New()
	rb := make([]byte, 32)
	_, err := rand.Read(rb)
	if err != nil {
		return fmt.Errorf("randomise bytes failed, %s", err)
	}
	s.Write(rb)
	switch fileType {
	case Manifest:
		fileName := fmt.Sprintf("%s-%x.json", Manifest.Strings(), s.Sum(nil))
		d.writeManifest(fileName)
	case MerkleTree:
		fileName := fmt.Sprintf("%s-%x.json", MerkleTree.Strings(), s.Sum(nil))
		d.writeMerkleTree(fileName)
	}
	return nil
}

func (d *DirectoryCrawler) writeManifest(fileName string) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	sigJson, _ := json.Marshal(d.Signatures)
	_, err = file.Write(sigJson)
	if err != nil {
		return
	}
}

func (d *DirectoryCrawler) writeMerkleTree(fileName string) {
	var homeTree hometree.MerkleTree
	home, merkleErr := homeTree.Build(d.FileDigests)
	if merkleErr.Err != nil {
		return
	}
	file, _ := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	sigJson, _ := json.Marshal(home)
	_, err := file.Write(sigJson)
	if err != nil {
		return
	}
}

func NewCrawler(root string, conf *Config) *DirectoryCrawler {
	return &DirectoryCrawler{
		Dir:  root,
		Conf: *conf}
}

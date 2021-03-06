package id3v2_test

import (
	asrt "assert"
	"io"
	"mp3agic/id3v2"
	"os"
	"testing"
)

const RES_DIR = "../../test-res/"

var assert = asrt.True

type BufReaderAt []byte

func (buf BufReaderAt) ReadAt(p []byte, off int64) (n int, err os.Error) {
	if off >= int64(len(buf)) {
		return 0, os.EOF
	}
	return copy(p, (buf)[off:]), nil
}

func bufWrap(pattern string) ([]byte, *io.SectionReader) {
	buf := make([]byte, len(pattern))
	copy(buf, pattern)
	return buf, io.NewSectionReader(BufReaderAt(buf), 0, int64(len(buf)))
}

const (
	id3v2_header = "ID3\x04\x00\x00\x00\x00\x02\x01"
)

func TestInitialiseFromHeaderBlockWithValidHeaders(t *testing.T) {
	buf, reader := bufWrap(id3v2_header)
	buf[3] = 2
	buf[4] = 0
	tag, err := id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Version() == "2.0", "expected version 2.0, got", tag.Version())

	buf[3] = 3
	tag, err = id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Version() == "3.0", "expected version 3.0, got", tag.Version())

	buf[3] = 4
	tag, err = id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Version() == "4.0", "expected version 4.0, got", tag.Version())
}

func TestCalculateCorrectDataLengthsFromHeaderBlock(t *testing.T) {
	buf, reader := bufWrap(id3v2_header)
	tag, err := id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.DataLength() == 257, "data length expected 257, got", tag.DataLength())

	buf[8] = 0x09
	buf[9] = 0x41
	tag, err = id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.DataLength() == 1217, "data length expected 1217, got", tag.DataLength())
}

func TestNonSupportedVersionInId3v2HeaderBlock(t *testing.T) {
	buf, reader := bufWrap(id3v2_header)
	buf[3] = 5
	buf[4] = 0
	_, err := id3v2.ExtractTag(reader)
	assert(t, err != nil, "expected error (wrong ID3v2 version), got nil")
}

func loadId3TagFile(fname string) (*id3v2.Tag, os.Error) {
	f, err := os.Open(RES_DIR+fname, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tag, err := id3v2.ExtractTag(f)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func TestReadFramesFromMp3With32Tag(t *testing.T) {
	tag, err := loadId3TagFile("v1andv23tags.mp3")
	if err != nil {
		t.Error("error loading file:", err)
		return
	}

	assert(t, tag.Version() == "3.0", "version expected 3.0, got", tag.Version())
	assert(t, tag.Length() == 0x44b, "length expected", 0x44b, "got", tag.Length())

	framesets := tag.FrameSets()
	if framesets == nil {
		t.Error("nil framesets")
		return
	}

	assert(t, len(framesets) == 12, "framesets length expected 12, got", len(framesets))
	assertFrameset := func(name string, count int) {
		fs, ok := framesets[name]
		if !ok {
			t.Error("absent frameset", name)
			return
		}
		n := len(fs)
		assert(t, n == count, "frameset", name, "elements expected", count, "got", n)
	}
	assertFrameset("TENC", 1)
	assertFrameset("WXXX", 1)
	assertFrameset("TCOP", 1)
	assertFrameset("TOPE", 1)
	assertFrameset("TCOM", 1)
	assertFrameset("COMM", 2)
	assertFrameset("TPE1", 1)
	assertFrameset("TALB", 1)
	assertFrameset("TRCK", 1)
	assertFrameset("TYER", 1)
	assertFrameset("TCON", 1)
	assertFrameset("TIT2", 1)
}

func TestReadId3v2WithFooter(t *testing.T) {
	tag, err := loadId3TagFile("v1andv24tags.mp3")
	if err != nil {
		t.Error("error loading file:", err)
		return
	}
	assert(t, tag.Version() == "4.0", "version expected 4.0, got", tag.Version())
	assert(t, tag.Length() == 0x44b, "length expected", 0x44b, "got", tag.Length())
}

func TestReadTagFieldsFromMp3With32tag(t *testing.T) {
	tag, err := loadId3TagFile("v1andv23tagswithalbumimage.mp3")
	if err != nil {
		t.Error("error loading file:", err)
		return
	}
	assert(t, tag.Track() == "1", "track expected 1, got", tag.Track())
	assert(t, tag.Artist() == "ARTIST123456789012345678901234", "artist", tag.Artist())
	assert(t, tag.Title() == "TITLE1234567890123456789012345", "title", tag.Title())
	assert(t, tag.Album() == "ALBUM1234567890123456789012345", "album", tag.Album())
	assert(t, tag.Year() == "2001", "year", tag.Year())
	assert(t, tag.Genre() == 0x0d, "genre expected", 0x0d, "got", tag.Genre())
	assert(t, tag.GenreDescription() == "Pop", "genre description", tag.GenreDescription())
	assert(t, tag.Comment() == "COMMENT123456789012345678901", "comment", tag.Comment())
	assert(t, tag.Composer() == "COMPOSER23456789012345678901234", "composer", tag.Composer())
	assert(t, tag.OriginalArtist() == "ORIGARTIST234567890123456789012", "original artist", tag.OriginalArtist())
	assert(t, tag.Copyright() == "COPYRIGHT2345678901234567890123", "copyright", tag.Copyright())
	assert(t, tag.Url() == "URL2345678901234567890123456789", "url", tag.Url())
	assert(t, tag.Encoder() == "ENCODER234567890123456789012345", "encoder", tag.Encoder())
	assert(t, len(tag.AlbumImage()) == 1885, "len(album image)", len(tag.AlbumImage()))
	assert(t, tag.AlbumImageMimeType() == "image/png", "album image mime type", tag.AlbumImageMimeType())
}

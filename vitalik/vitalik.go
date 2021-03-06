//package vitalik
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	//err := SplitFile("storedFile", 1024*1024, 30, 10)
	//if err != nil {
	//	fmt.Println(err)
	//}
	pathToChunks := []string{"Chunk0", "Chunk1", "Chunk2", "Chunk3", "Chunk4",
		"Chunk5", "Chunk6", "Chunk7", "Chunk8", "Chunk9"}
	chunkIds := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	err := RetrieveFile(pathToChunks, chunkIds, 10485765)
	if err != nil {
		fmt.Println(err)
	}
}

func SplitFile(filePath string, chunkSizeBytes int, totChunk int, minChunk int) error {
	/*
		Generates chunks derived from the file, to be distributed each to a different node.
			filePath		path to the file to be split
			chunkSizeByts	size in bytes of each chunk
			totChunk		total number of chunks. The sum of the sizes of every
							chunk may exceed the size of the initial file, because
							of redundancy
			minChunk		minimal number of chunks needed in order to rebuild the file
	*/
	initialFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer initialFile.Close()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// check for bad arguments
	if int64(chunkSizeBytes*totChunk) < fileInfo.Size() {
		return fmt.Errorf("SplitFile: Total size of chunks is smaller than the file")
	} else if minChunk > totChunk {
		return fmt.Errorf("SplitFile: minChunk must be smaller than totChunk")
	}

	// Open one file for each chunk
	chunkWriters := make([]*os.File, totChunk)
	for i := 0; i < totChunk; i++ {
		fileName := fmt.Sprintf("Chunk%d", i)
		chunkWriters[i], err = os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0660) // Open files as write only
		if err != nil {
			return err
		}
		defer chunkWriters[i].Close()
	}
	// Create buffers and counters
	bytesToWrite := 100
	chunkBuffers := make([][]byte, totChunk)
	for i := range chunkBuffers {
		chunkBuffers[i] = make([]byte, bytesToWrite)
	}
	buff := make([]byte, minChunk*bytesToWrite)
	workCounter := 0
	bytesWrote := 0
	//Fill the x with natural numbers
	xCoords := make([]byte, minChunk)
	for i := range xCoords {
		xCoords[i] = byte(i)
	}

	// Main Loop
	for term := 1; term != 0; {
		n, err := initialFile.Read(buff)
		if err != nil {
			if err == io.EOF {
				term = 0    // end loop next time
				if n == 0 { // It means that the file size is a multiple of min Chunk.
					return nil // Terminates the function. Everything has been done
				}
			} else {
				return err
			}
		}
		// A useful Counter of the work done
		workCounter += 1
		if workCounter%(100000/bytesToWrite) == 0 { //Display a message with the progress
			fmt.Printf("%d bytes read\n", workCounter*minChunk*bytesToWrite)
		}

		for i := 0; i*minChunk < n; i++ {
			pointNumber := minChunk // Number of points to interpolate
			if (i+1)*minChunk > n {
				pointNumber = n % minChunk
			}
			bytesWrote = i + 1
			// Adjust x Coords
			xCoords = xCoords[:pointNumber]
			//Fill the y with the bytes read
			yCoords := buff[minChunk*i : minChunk*i+pointNumber]
			// Interpolate
			points, err := FiniteFieldLagrangeInterpolation(xCoords, yCoords, totChunk)
			if err != nil {
				return err
			}
			//Write into the buffers
			for j := 0; j < totChunk; j++ {
				chunkBuffers[j][i] = points[j]
			}
		}
		// Write into the files
		for j := range chunkBuffers {
			_, err = chunkWriters[j].Write(chunkBuffers[j][:bytesWrote])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RetrieveFile(pathToChunks []string, chunkIds []byte, fileSize int64) error {
	/*
	   Rebuilds the file, given a number of chunks generated with the function SplitFile.

	   pathToChunks    list containing the paths of the minimun number of chunks necessary.
	   chunkIds        the id of every chunk
	   fileSize        the size of the file to rebuild.
	*/

	// Check for bad arguments
	minChunk := len(pathToChunks)
	if minChunk != len(chunkIds) {
		return fmt.Errorf("RetrieveFile: %d paths given, but %d ids given. They must be the same number.",
			minChunk, len(chunkIds))
	}
	chunkInfo, err := os.Stat(pathToChunks[0])
	if err != nil {
		return err
	}
	chunkSize := chunkInfo.Size()
	for _, path := range pathToChunks {
		chunkInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		if chunkSize != chunkInfo.Size() {
			return fmt.Errorf("RetrieveFile: All chunk files must be the same size")
		}
	}
	if int64(minChunk)*chunkSize < fileSize {
		return fmt.Errorf("RetrieveFile: total size of chunks is smaller than fileSize.")
	}
	// Open Chunks
	chunkReader := make([]*os.File, minChunk)
	for i := range chunkReader {
		chunkReader[i], err = os.Open(pathToChunks[i])
		if err != nil {
			return err
		}
		defer chunkReader[i].Close()
	}
	// Create RestoredFile
	restoredFile, err := os.OpenFile("restoredFile", os.O_WRONLY|os.O_CREATE, 0660) // Open file as write only
	if err != nil {
		return err
	}
	// Create Buffers and Counters
	bytesToRead := 100
	chunkBuffers := make([][]byte, minChunk)
	for i := range chunkBuffers {
		chunkBuffers[i] = make([]byte, bytesToRead)
	}
	buff := make([]byte, minChunk*bytesToRead)
	workCounter := 0
	bytesWrote := 0

	// Main Loop
	for term := 1; term != 0; {
		var n int
		for i := range chunkBuffers {
			n, err = chunkReader[i].Read(chunkBuffers[i])
		}
		if err == io.EOF || n != bytesToRead {
			term = 0    // end loop next time
			if n == 0 { // This means that chunkSize is a multiple of bytesToRead.
				return nil //Terminates the funcion. Everything has been done.
			}
			if n != bytesToRead {
				fmt.Println(err, n, term)
			}
		} else if err != nil {
			return err
		}

		// A useful Counter of the work done
		workCounter += 1
		if workCounter%(100000/bytesToRead) == 0 { //Display a message with the progress
			fmt.Printf("%d bytes written\n", workCounter*minChunk*bytesToRead)
		}

		for i := 0; i < n; i++ {
			pointNumber := minChunk // Number of points to interpolate
			if n != bytesToRead && i == n-1 {
				pointNumber = int((int64(minChunk) * chunkSize) - fileSize)
			}
			bytesWrote = i*minChunk + pointNumber
			//Fill the x coords with chunk ids
			xCoords := chunkIds[:pointNumber]
			//Fill the y coords with the bytes read
			yCoords := make([]byte, pointNumber)
			for j := range yCoords {
				yCoords[j] = chunkBuffers[j][i]
			}
			// Interpolate
			points, err := FiniteFieldLagrangeInterpolation(xCoords, yCoords, pointNumber)
			if err != nil {
				return err
			}
			//write to the buffer
			for j := 0; j < pointNumber; j++ {
				buff[i*minChunk+j] = points[j]
			}
		}
		// Write to the file
		_, err = restoredFile.Write(buff[:bytesWrote])
		if n != bytesToRead {
			fmt.Println(buff[:bytesWrote], term)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func FiniteFieldLagrangeInterpolation(xCoords []byte, yCoords []byte, pointNumber int) ([]byte, error) {
	/*
		Interpolates a polynomial given a set of points inside a finite field
		of order 256.

			xCoords			the values of the x coordinates of each point
			yCoords			the corresponding value of y. Being bytes, 0<=y<=255
			pointNumber		the number of points to be interpolated.

		xCoords and yCoords must have the same length.
		The degree of the polynomial is supposed to be the length of the parameters.
		Returns a slice of pointNumber elements, where element i contains the value y
		of the polynomial at x=i.

		Inside the GF(256) (Galois Field with 256 elements), the arithmetic operations
		are carried between single bytes in the following way:
		a + b -> a ^ b
		a - b -> a ^ b
		a * b -> GFMul(a,b)
		a / b -> GFDiv(a,b)
	*/
	var degree int = len(xCoords)
	if degree != len(yCoords) {
		return nil, fmt.Errorf("Lagrange: %d x coords given, but %d y coords. They must be the same number",
			degree, len(yCoords))
	}
	// Check for repeated elements, that would screw up Lagrange interpolation
	for i := 0; i < degree; i++ {
		for j := 0; j < degree; j++ {
			if i != j && xCoords[i] == xCoords[j] {
				return nil, fmt.Errorf("Lagrange: All the x coordinates must be different")
			}
		}
	}

	var numerator byte
	var denominator byte
	var y byte
	var div byte

	result := make([]byte, pointNumber)

	// Do the Math
	for x := 0; x < pointNumber; x++ {
		y = 0
		for i := 0; i < degree; i++ {
			numerator = 0x01
			denominator = 0x01

			for j := 0; j < degree; j++ {
				if i != j {
					numerator = GFMul(numerator, (byte(x) ^ xCoords[j]))
					denominator = GFMul(denominator, (xCoords[i] ^ xCoords[j]))
				}
			}
			div, _ = GFDiv(numerator, denominator)
			y = y ^ (GFMul(div, yCoords[i]))
		}
		result[x] = y
	}
	return result, nil
}

func GFMul(a byte, b byte) byte {
	/*	Performs the multiplication of the two polynomials a and b, which have
		coefficients in GF(2) (e.g. either 0 or 1) and lye in GF(256) (e.g. have
		at most degree 7). The multiplycation is easily performed using the
		exponential and logarithmic tables of the generator polynomial x + 1.
	*/
	if a == 0 || b == 0 {
		return 0x00
	}

	return byte(expTable[(int(logTable[a])+int(logTable[b]))%0xff])
}

func GFDiv(a byte, b byte) (byte, error) {
	// Same as FG mul, but performs a division
	if b == 0 {
		return 0x00, fmt.Errorf("GFDiv: Tried to do %d/%d. Cannot divide by zero", a, b)
	}
	if a == 0 {
		return 0x00, nil
	}
	return byte(expTable[((int(logTable[a])-int(logTable[b]))%0xff+0xff)%0xff]), nil
}

func GFMulPeasant(a byte, b byte) byte {
	/* Performs the multiplicaition of the two polynomials a and b, with
	   coefficients in GF(2) and over GF(256).
	   This Galois Field is obtained from the remainder classes of the polynomial
	   x^8 + x^4 + x^3 + x + 1, which is irreducible (prime) in GF(2).
	   Therefore, and being of degree 8, it generates 2^8=256 polynomials or
	   remainder classes.
	   The polynomial can also be expresses as 100011011 = 0x11b.
	   The multilplication is performed using the peasant algorithm
	   https://en.wikipedia.org/wiki/Finite_field_arithmetic#Multiplication
	*/
	var p byte = 0 // the final product
	for b > 0 {
		if (b & 1) != 0 { // if b is odd
			p ^= a // p = p+a. In GF(2) the sum is equivalent to xor
		}
		if (a & 0x80) != 0 { // if a > 128
			a = (a << 1) ^ 0x1b // which is 0x11b, without the leftmost bit, because
			// otherwise it would cause overflow.
		} else {
			a <<= 1 // a*2
		}
		b >>= 1 // b/2
	}
	return p
}

/*
The following two tables are needed to simplify the calculation of multiplications and
divisions inside the GF(256). They are created such as expTable[i] = 0x03^i
and logTable[i] = log_0x03(i). 0x03, being the polynomial x+1, is the simplest generator
of the field. This means that the first 256 powers of 0x03 will assume each polynomial of
the field.
The tables have been generated using the GFMulPeasant function.
For more information http://www.cs.utsa.edu/~wagner/laws/FFM.html
WARNING: In logTable, the 0th element is set to 0x00, while it is actually not defined,
	because it represents log_0x03(0). Therefore it should never be used.
*/
var expTable = [256]byte{
	0x01, 0x03, 0x05, 0x0F, 0x11, 0x33, 0x55, 0xFF, 0x1A, 0x2E, 0x72, 0x96, 0xA1, 0xF8, 0x13, 0x35,
	0x5F, 0xE1, 0x38, 0x48, 0xD8, 0x73, 0x95, 0xA4, 0xF7, 0x02, 0x06, 0x0A, 0x1E, 0x22, 0x66, 0xAA,
	0xE5, 0x34, 0x5C, 0xE4, 0x37, 0x59, 0xEB, 0x26, 0x6A, 0xBE, 0xD9, 0x70, 0x90, 0xAB, 0xE6, 0x31,
	0x53, 0xF5, 0x04, 0x0C, 0x14, 0x3C, 0x44, 0xCC, 0x4F, 0xD1, 0x68, 0xB8, 0xD3, 0x6E, 0xB2, 0xCD,
	0x4C, 0xD4, 0x67, 0xA9, 0xE0, 0x3B, 0x4D, 0xD7, 0x62, 0xA6, 0xF1, 0x08, 0x18, 0x28, 0x78, 0x88,
	0x83, 0x9E, 0xB9, 0xD0, 0x6B, 0xBD, 0xDC, 0x7F, 0x81, 0x98, 0xB3, 0xCE, 0x49, 0xDB, 0x76, 0x9A,
	0xB5, 0xC4, 0x57, 0xF9, 0x10, 0x30, 0x50, 0xF0, 0x0B, 0x1D, 0x27, 0x69, 0xBB, 0xD6, 0x61, 0xA3,
	0xFE, 0x19, 0x2B, 0x7D, 0x87, 0x92, 0xAD, 0xEC, 0x2F, 0x71, 0x93, 0xAE, 0xE9, 0x20, 0x60, 0xA0,
	0xFB, 0x16, 0x3A, 0x4E, 0xD2, 0x6D, 0xB7, 0xC2, 0x5D, 0xE7, 0x32, 0x56, 0xFA, 0x15, 0x3F, 0x41,
	0xC3, 0x5E, 0xE2, 0x3D, 0x47, 0xC9, 0x40, 0xC0, 0x5B, 0xED, 0x2C, 0x74, 0x9C, 0xBF, 0xDA, 0x75,
	0x9F, 0xBA, 0xD5, 0x64, 0xAC, 0xEF, 0x2A, 0x7E, 0x82, 0x9D, 0xBC, 0xDF, 0x7A, 0x8E, 0x89, 0x80,
	0x9B, 0xB6, 0xC1, 0x58, 0xE8, 0x23, 0x65, 0xAF, 0xEA, 0x25, 0x6F, 0xB1, 0xC8, 0x43, 0xC5, 0x54,
	0xFC, 0x1F, 0x21, 0x63, 0xA5, 0xF4, 0x07, 0x09, 0x1B, 0x2D, 0x77, 0x99, 0xB0, 0xCB, 0x46, 0xCA,
	0x45, 0xCF, 0x4A, 0xDE, 0x79, 0x8B, 0x86, 0x91, 0xA8, 0xE3, 0x3E, 0x42, 0xC6, 0x51, 0xF3, 0x0E,
	0x12, 0x36, 0x5A, 0xEE, 0x29, 0x7B, 0x8D, 0x8C, 0x8F, 0x8A, 0x85, 0x94, 0xA7, 0xF2, 0x0D, 0x17,
	0x39, 0x4B, 0xDD, 0x7C, 0x84, 0x97, 0xA2, 0xFD, 0x1C, 0x24, 0x6C, 0xB4, 0xC7, 0x52, 0xF6, 0x01,
}

var logTable = [256]byte{
	0x00, 0x00, 0x19, 0x01, 0x32, 0x02, 0x1A, 0xC6, 0x4B, 0xC7, 0x1B, 0x68, 0x33, 0xEE, 0xDF, 0x03,
	0x64, 0x04, 0xE0, 0x0E, 0x34, 0x8D, 0x81, 0xEF, 0x4C, 0x71, 0x08, 0xC8, 0xF8, 0x69, 0x1C, 0xC1,
	0x7D, 0xC2, 0x1D, 0xB5, 0xF9, 0xB9, 0x27, 0x6A, 0x4D, 0xE4, 0xA6, 0x72, 0x9A, 0xC9, 0x09, 0x78,
	0x65, 0x2F, 0x8A, 0x05, 0x21, 0x0F, 0xE1, 0x24, 0x12, 0xF0, 0x82, 0x45, 0x35, 0x93, 0xDA, 0x8E,
	0x96, 0x8F, 0xDB, 0xBD, 0x36, 0xD0, 0xCE, 0x94, 0x13, 0x5C, 0xD2, 0xF1, 0x40, 0x46, 0x83, 0x38,
	0x66, 0xDD, 0xFD, 0x30, 0xBF, 0x06, 0x8B, 0x62, 0xB3, 0x25, 0xE2, 0x98, 0x22, 0x88, 0x91, 0x10,
	0x7E, 0x6E, 0x48, 0xC3, 0xA3, 0xB6, 0x1E, 0x42, 0x3A, 0x6B, 0x28, 0x54, 0xFA, 0x85, 0x3D, 0xBA,
	0x2B, 0x79, 0x0A, 0x15, 0x9B, 0x9F, 0x5E, 0xCA, 0x4E, 0xD4, 0xAC, 0xE5, 0xF3, 0x73, 0xA7, 0x57,
	0xAF, 0x58, 0xA8, 0x50, 0xF4, 0xEA, 0xD6, 0x74, 0x4F, 0xAE, 0xE9, 0xD5, 0xE7, 0xE6, 0xAD, 0xE8,
	0x2C, 0xD7, 0x75, 0x7A, 0xEB, 0x16, 0x0B, 0xF5, 0x59, 0xCB, 0x5F, 0xB0, 0x9C, 0xA9, 0x51, 0xA0,
	0x7F, 0x0C, 0xF6, 0x6F, 0x17, 0xC4, 0x49, 0xEC, 0xD8, 0x43, 0x1F, 0x2D, 0xA4, 0x76, 0x7B, 0xB7,
	0xCC, 0xBB, 0x3E, 0x5A, 0xFB, 0x60, 0xB1, 0x86, 0x3B, 0x52, 0xA1, 0x6C, 0xAA, 0x55, 0x29, 0x9D,
	0x97, 0xB2, 0x87, 0x90, 0x61, 0xBE, 0xDC, 0xFC, 0xBC, 0x95, 0xCF, 0xCD, 0x37, 0x3F, 0x5B, 0xD1,
	0x53, 0x39, 0x84, 0x3C, 0x41, 0xA2, 0x6D, 0x47, 0x14, 0x2A, 0x9E, 0x5D, 0x56, 0xF2, 0xD3, 0xAB,
	0x44, 0x11, 0x92, 0xD9, 0x23, 0x20, 0x2E, 0x89, 0xB4, 0x7C, 0xB8, 0x26, 0x77, 0x99, 0xE3, 0xA5,
	0x67, 0x4A, 0xED, 0xDE, 0xC5, 0x31, 0xFE, 0x18, 0x0D, 0x63, 0x8C, 0x80, 0xC0, 0xF7, 0x70, 0x07,
}

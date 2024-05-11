<<<<<<<<<<<<<<  ✨ Codeium Command 🌟 >>>>>>>>>>>>>>>>
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

+// Constants for block size, number of workers, and match channel size
const (
+	blockSize     = 4 * 1024 * 1024 // Block size in bytes
+	numWorkers    = 8               // Number of workers
+	matchChanSize = 1024            // Match channel size
-	blockSize     = 4 * 1024 * 1024 // tamaño del bloque en bytes
-	numWorkers    = 8               // número de trabajadores
-	matchChanSize = 1024            // tamaño del canal de coincidencias
)

+// Main function of the program
-// Función principal del programa
func main() {
+	// Parse command line arguments
+	recursive := flag.Bool("recursive", false, "Perform recursive search")
+	inputPath := flag.String("path", ".", "Input directory path")
+	pattern := flag.String("pattern", "", "Regular expression pattern for matching")
+	outputFile := flag.String("output", "results.txt", "Output file path")
-	// Parseo de argumentos de la línea de comandos
-	recursive := flag.Bool("recursive", false, "Realizar una búsqueda recursiva")
-	inputPath := flag.String("path", ".", "Ruta del directorio de entrada")
-	pattern := flag.String("pattern", "", "Expresión regular para buscar coincidencias")
-	outputFile := flag.String("output", "results.txt", "Ruta del archivo de salida")
	flag.Parse()

+	// Check if regular expression pattern is provided
-	// Verificación de que se proporcionó una expresión regular
	if *pattern == "" {
		flag.Usage()
		return
	}

+	// Compile the regular expression
-	// Compilación de la expresión regular
	regex, err := regexp.CompilePOSIX(*pattern)
	if err != nil {
+		fmt.Printf("Error compiling regular expression: %v\n", err)
-		fmt.Printf("Error al compilar la expresión regular: %v\n", err)
		return
	}

+	// Open the output file
-	// Apertura del archivo de salida
	output, err := os.Create(*outputFile)
	if err != nil {
+		fmt.Printf("Error creating output file: %v\n", err)
-		fmt.Printf("Error al crear el archivo de salida: %v\n", err)
		return
	}
	defer output.Close()

+	// Create channels for block processing and match writing
-	// Creación de canales para el procesamiento de bloques y coincidencias
	blockChan := make(chan []byte, numWorkers)
	matchChan := make(chan []byte, matchChanSize)
	errChan := make(chan error, numWorkers)
	done := make(chan struct{})
	var wg sync.WaitGroup

+	// Function to print the name of the finished file
-	// Función para imprimir el nombre del archivo terminado
	printFileName := func(filename string) {
+		fmt.Printf("File finished: %s\n", filename)
-		fmt.Printf("Archivo terminado: %s\n", filename)
	}

+	// Create workers
-	// Creación de trabajadores
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(regex, blockChan, matchChan, errChan, &wg, done)
	}

+	// Process the input path
-	// Procesamiento de la ruta de entrada
	go func() {
		defer close(blockChan)
		err := processPath(*inputPath, *recursive, blockChan, errChan, printFileName)
		if err != nil {
+			fmt.Printf("Error processing input path: %v\n", err)
-			fmt.Printf("Error al procesar la ruta de entrada: %v\n", err)
		}
	}()

+	// Close the match channel when all blocks are processed
-	// Cierre del canal de coincidencias cuando se hayan procesado todos los bloques
	go func() {
		wg.Wait()
		close(matchChan)
		close(done)
	}()

+	// Write matches to the output file
-	// Escritura de las coincidencias en el archivo de salida
	writer := bufio.NewWriter(output)
	for match := range matchChan {
		_, err := writer.Write(match)
		if err != nil {
+			fmt.Printf("Error writing to output file: %v\n", err)
-			fmt.Printf("Error al escribir en el archivo de salida: %v\n", err)
			break
		}
		_, err = writer.WriteString("\n")
		if err != nil {
+			fmt.Printf("Error writing to output file: %v\n", err)
-			fmt.Printf("Error al escribir en el archivo de salida: %v\n", err)
			break
		}
	}

+	// Print errors if any
-	// Impresión de errores si los hubo
	if len(errChan) > 0 {
+		fmt.Println("The following errors occurred during processing:")
-		fmt.Println("Se produjeron los siguientes errores durante el procesamiento:")
		for err := range errChan {
			fmt.Println(err)
		}
	}
	fmt.Println("Finish bro! -SrFvchx")
	writer.Flush()
}

+// Process the input path and send blocks through the block channel
+//
+// Parameters:
+// - path: input path to search
+// - recursive: perform recursive search
+// - blockChan: channel to send blocks
+// - errChan: channel to send errors
+// - printFileName: function to print the name of the finished file
+//
+// Returns:
+// - error: error occurred during processing
-// Procesa la ruta de entrada y envía los bloques a través del canal de bloques
func processPath(path string, recursive bool, blockChan chan<- []byte, errChan chan<- error, printFileName func(string)) error {
	if path == "here" {
		path, _ = os.Getwd()
	}
	var filesProcessed int
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			errChan <- fmt.Errorf("Error al recorrer el directorio %s: %v", p, err)
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(d.Name(), ".txt") || strings.HasSuffix(d.Name(), ".csv") || strings.HasSuffix(d.Name(), ".sql")) {
			err := processFile(p, blockChan, d.Name())
			if err != nil {
				errChan <- fmt.Errorf("Error al procesar el archivo %s: %v", p, err)
			}
			filesProcessed++
			printFileName(d.Name())
		}
		return nil
	})
	if err != nil {
		errChan <- fmt.Errorf("Error al recorrer el directorio %s: %v", path, err)
		return err
	}
	fmt.Printf("Total de archivos procesados: %d\n", filesProcessed)
	return nil
}

// Lee el archivo y envía los bloques a través del canal de bloques
func processFile(file string, blockChan chan<- []byte, filename string) error {
	fd, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Error al abrir el archivo %s: %v", file, err)
	}
	defer fd.Close()
	buf := make([]byte, blockSize)
	blockHeader := []byte(fmt.Sprintf("#################################################\n%s\n\n", filename))
	blockChan <- blockHeader
	for {
		n, err := fd.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Error al leer el archivo %s: %v", file, err)
		}
		if n > 0 {
			blockChan <- buf[:n]
		}
	}
	return nil
}

// Procesa los bloques y envía las coincidencias a través del canal de coincidencias
func worker(regex *regexp.Regexp, blockChan <-chan []byte, matchChan chan<- []byte, errChan chan<- error, wg *sync.WaitGroup, done <-chan struct{}) {
	defer wg.Done()
	for {
		select {
		case block, ok := <-blockChan:
			if !ok {
				return
			}
			matches := regex.FindAll(block, -1)
			for _, match := range matches {
				matchChan <- match
			}
		case <-done:
			return
		}
	}
}
<<<<<<<  cf0a7477-f203-4368-a8f0-787720daba24  >>>>>>>
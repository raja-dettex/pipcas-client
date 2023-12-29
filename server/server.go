package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/raja-dettex/pipcas-client/service"
	"github.com/raja-dettex/pipcas-client/storage"
)

type ServerOpts struct {
	ListenAddr string
	PipCasHost string
	PipCasPort string
}

type APIServer struct {
	opts        ServerOpts
	storage     storage.FileStorage
	DB          service.PipCasDbService
	errCh       chan error
	connectedCh chan bool
}

func NewAPIServer(opts ServerOpts, storage storage.FileStorage, db service.PipCasDbService) *APIServer {
	return &APIServer{opts: opts, storage: storage, DB: db, errCh: make(chan error), connectedCh: make(chan bool)}
}

func (s *APIServer) Start() error {
	go s.storage.Evict()
	fmt.Println("server is yet to start")
	for {
		// Attempt to create the schema
		err := s.DB.CreateSchema()
		fmt.Println("error", err)
		if err != nil {
			// If an error occurs, handle it and then retry after a delay
			fmt.Println("Error creating schema:", err)
			time.Sleep(time.Second * 4) // Add a delay before retrying

		} else {
			// If successful, signal that the schema is created and proceed
			//s.connectedCh <- true
			fmt.Println("server is starting")
			break // Exit the loop if successful
		}
	}
	s.handlers()
	return http.ListenAndServe(s.opts.ListenAddr, nil)

}

func (s *APIServer) address() string {
	return fmt.Sprintf("%s:%s", s.opts.PipCasHost, s.opts.PipCasPort)
}

func (s *APIServer) handlers() {
	http.HandleFunc("/upload", s.UploadFileHandler)
	http.HandleFunc("/download/", s.DownloadFileHandler)
}

func (s *APIServer) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	var (
		readBuff  = make([]byte, 2048)
		writeBuff = make([]byte, 5432)
	)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n, err := file.Read(writeBuff)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("File content", string(writeBuff[:n]))
	defer file.Close()
	// send the file content to storage server
	conn, err := net.Dial("tcp", s.address())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = conn.Write([]byte(fmt.Sprintf("write %s %s", handler.Filename, string(writeBuff[:n]))))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		n, err := conn.Read(readBuff)
		if err == io.EOF {
			fmt.Println("end of line closing connection")
			conn.Close()
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			conn.Close()
			return
		}
		storageAddress := string(readBuff[:n])
		fmt.Println(storageAddress)
		err = s.DB.Save(storageAddress, handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lFile, err := s.storage.Write(handler.Filename, bytes.NewBuffer(writeBuff[:n]))
		if err != nil {
			fmt.Println("error writing to local file storage ", err)
		}
		defer lFile.Close()
		fmt.Fprintf(w, "response: storage location:  %s", storageAddress)
		conn.Close()
	}
}

func (s *APIServer) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	var (
		readBuff  = make([]byte, 2048)
		writeBuff = make([]byte, 5432)
	)
	fileName := r.URL.Path[len("/download/"):]
	file, err := s.storage.Read(fileName)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			storageAddr, err := s.DB.GetBy(fileName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			conn, err := net.Dial("tcp", s.address())

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer conn.Close()
			fmt.Println(storageAddr)
			_, err = conn.Write([]byte(fmt.Sprintf("read %s %s", fileName, storageAddr)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer conn.Close()
			for {
				n, err := conn.Read(readBuff)
				if err == io.EOF {
					fmt.Println("end of line closing connection")
					return
				}
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				//fmt.Fprintf(w, "response %s", string(readBuff[:n]))
				file, err := s.storage.Write(fileName, bytes.NewBuffer(readBuff[:n]))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				defer file.Close()
				fileStat, err := file.Stat()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Length", fmt.Sprintf("%d", fileStat.Size()))

				// Copy the file to the response writer
				http.ServeContent(w, r, fileName, fileStat.ModTime(), file)
			}

		} else {
			fmt.Fprintf(w, err.Error(), http.StatusBadRequest)
		}
	}
	n, err := file.Read(writeBuff)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("buffer", writeBuff[:n])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileStat.Size()))

	// Copy the file to the response writer
	http.ServeContent(w, r, fileName, fileStat.ModTime(), file)
}

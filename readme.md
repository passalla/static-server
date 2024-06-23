# Static File Server

Proyek ini adalah server HTTP sederhana yang melayani file statis menggunakan framework Gin. Server ini dapat dikonfigurasi melalui argumen command line.

## Cara Menggunakan

1. Clone repositori ini:

   ```sh
   git clone https://github.com/joco-ding/static-server.git
   cd static-server
   ```

2. Compile kode Go:

   ```sh
   go build -o static-file-server cmd/main.go
   ```

3. Jalankan server dengan opsi yang diinginkan:

   ```sh
   ./static-file-server [options]
   ```

## Opsi

- `-port`: Port untuk menjalankan server. Default: `8080`
- `-static`: Direktori untuk file statis. Default: `./`
- `-debug`: Mode debug. Default: `false`

## Contoh

Menjalankan server di port 9090, melayani file statis dari direktori `public`, dan mengaktifkan mode debug:

```sh
./static-file-server -port 9090 -static ./public -debug
```

## Struktur Direktori

Berikut adalah struktur direktori yang direkomendasikan untuk proyek ini:

```
.
├── main.go
├── README.md
└── public
    ├── index.html
    └── ...
```

- `main.go`: File sumber utama untuk server.
- `README.md`: File ini.
- `public`: Direktori tempat menyimpan file statis Anda (misalnya, `index.html`, CSS, JS).

## Lisensi

Proyek ini dilisensikan di bawah [MIT License](LICENSE).
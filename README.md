<h1 align="center"><i>Todolist</i></h1>

<div align="center">

[![CI](https://github.com/Lai-YT/todolist/actions/workflows/ci.yml/badge.svg)](https://github.com/Lai-YT/todolist/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/Lai-YT/todolist/graph/badge.svg?token=VMJZ7U806J)](https://codecov.io/gh/Lai-YT/todolist)
[![GitHub Issues Open](https://img.shields.io/github/issues/Lai-YT/todolist.svg)](https://github.com/Lai-YT/todolist/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/Lai-YT/todolist.svg)](https://github.com/Lai-YT/todolist/pulls)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

</div>

---

<p align="center">
    Todolist is a simple todo list API server,
    <br>
    allowing you to add, remove, and mark tasks as done.
</p>

## Brief

This project is a simple endeavor to learn _Go_, following the blog post [Build a Todolist API Server in Golang](https://www.fadhil-blog.dev/blog/golang-todolist/) with a few modifications.

The API server uses:

- [MySQL](https://www.mysql.com/) as our database
- [GORM](https://gorm.io/index.html) as the ORM to interact with our database
- Request routing using [gorilla/mux](https://github.com/gorilla/mux)
- [Logrus](https://github.com/sirupsen/logrus) for logging

There's also a frontend for this project, which was initially created by [themaxsandelin](https://github.com/themaxsandelin), modified by [sdil](https://github.com/sdil), and finally tailored by me. You can find it at [Lai-YT/todolist-frontend](https://github.com/Lai-YT/todolist-frontend).

## License

Todolist is licensed under the [MIT license](LICENSE).

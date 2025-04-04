# gorgx

This repository contains a simple implementation of a regular expression engine in Go.
The goal of this project is to build a custom engine that can validate for example
email addresses using the pattern `[a-zA-Z][a-zA-Z0-9_.]+@[a-zA-Z0-9]+.[a-zA-Z]{2,}`.
This regex pattern checks if an email follows a common
format such as `user@example.com`.

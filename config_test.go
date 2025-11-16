package main

import "os"

func init() {
    os.Setenv("DB_HOST", "localhost")
    os.Setenv("DB_PORT", "5432") 
    os.Setenv("DB_USER", "postgres")
    os.Setenv("DB_PASSWORD", "password")
    os.Setenv("DB_NAME", "pr_reviewer")
}
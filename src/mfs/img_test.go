package mfs

import (
  "testing"
)

func TestImg(t *testing.T) {
  img := NewImg("img")
  if img == nil {
    t.Fatal("NewImg Error")
  }

  if !img.LoadSuper() {
    t.Fatal("Load Super Error")
  }
  img.PutObj(123, nil)
}

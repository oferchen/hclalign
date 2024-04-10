// example2.tf
variable "example-var1" {
  description = "An example variable"
}
variable "example-var2" {
  type        = string
  description = "An example variable"
}

variable "example-var3" {
  type = string
  default = {
    "key1" = "value1"
  }
  description = "An example variable"
}

variable "example-var4" {
  type = object
  default = {
    object1 = {
      key1 = "value1"
      key2 = "value2"
      key3 = "value3"
      key4 = "value4"
      key5 = "value5"
    }
  }
  description = "An example variable"
}

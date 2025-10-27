variable "name" {
  description = "Name of database instance"
}

variable "region" {
  description = "The region for the database instance"
}

variable "size" {
  description = "the size of the database instance"
  default     = "db-f1-micro"
}

variable "db_version" {
  description = "The version of the database"
  default     = "POSTGRES_15"
}

output "values" {
  value = {
    name       = var.name
    region     = var.region
    size       = var.size
    db_version = var.db_version
  }
}
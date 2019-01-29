variable "aws_access_key" {
  description = "AWS access key."
}

variable "aws_secret_key" {
  description = "AWS secret key."
}

variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "us-west-2"
}

//variable "aws_account_id" {
//  description = "AWS account ID"
//}

variable "unique_s3_bucket_prefix" {
  description = "Unique prefix for s3 bucket"
}

variable "az_count" {
  description = "Number of AZs to cover in a given AWS region"
  default     = "2"
}

variable "app_image" {
  description = "Docker image to run in the ECS cluster"
  default     = "ericlien/hostname-docker:latest"
}

variable "app_port" {
  description = "Port exposed by the docker image to redirect traffic to"
  default     = 8080
}

variable "app_count" {
  description = "Number of docker containers to run"
  default     = 1
}

variable "fargate_cpu" {
  description = "Fargate instance CPU units to provision (1 vCPU = 1024 CPU units)"
  default     = "256"
}

variable "fargate_memory" {
  description = "Fargate instance memory to provision (in MiB)"
  default     = "512"
}

variable "dynamo_read_capacity" {
  description = "Dynamodb table read capacity"
  default = 5
}

variable "dynamo_write_capacity" {
  description = "Dynamodb table write capacity"
  default = 5
}

variable "environment" {
  description = "Environment of development, testing, staging, production"
  default = "production"
}
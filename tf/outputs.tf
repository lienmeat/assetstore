output "alb_hostname" {
  value = "${aws_alb.main.dns_name}"
}

output "myapp-repo" {
  value = "${aws_ecr_repository.repo.repository_url}"
}
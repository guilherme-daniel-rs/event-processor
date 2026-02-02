provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

locals {
  main_queue_name      = "events-main"
  dlq_queue_name       = "events-dlq"
  max_receive_count    = 5
}

resource "aws_sqs_queue" "dlq" {
  name = local.dlq_queue_name
  message_retention_seconds = 1209600 # 14 dias
}

resource "aws_sqs_queue" "main" {
  name = local.main_queue_name

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = local.max_receive_count
  })

  receive_wait_time_seconds = 20
}

output "main_queue_url" {
  value = aws_sqs_queue.main.url
}

output "dlq_queue_url" {
  value = aws_sqs_queue.dlq.url
}

output "dlq_queue_arn" {
  value = aws_sqs_queue.dlq.arn
}

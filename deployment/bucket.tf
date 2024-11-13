// standard bucket
resource "aws_s3_bucket" "csv_txns_files" {
  tags = {
    Name = "Transaction CSV files"
  }
  force_destroy = true
}

// trigger lambda
resource "aws_s3_bucket_notification" "bucket_terraform_notification" {
  bucket = aws_s3_bucket.csv_txns_files.id
  lambda_function {
    lambda_function_arn = aws_lambda_function.function.arn
    events = ["s3:ObjectCreated:*"]
  }
  depends_on = [ aws_lambda_permission.allow_terraform_bucket ]
}

// Simple output to easy debugging
output "csv_txns_files_bucket" {
  value = aws_s3_bucket.csv_txns_files.id
}
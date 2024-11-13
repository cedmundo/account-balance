// allow lambda service to assume (use) the role with such policy
data "aws_iam_policy_document" "assume_lambda_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

// create lambda role, that lambda function can assume (use)
resource "aws_iam_role" "lambda" {
  name               = "AssumeLambdaRole"
  description        = "Role for lambda to assume lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role.json
}

data "aws_iam_policy_document" "allow_lambda_logging" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "arn:aws:logs:*:*:*",
    ]
  }
}

// create a policy to allow writing into logs and create logs stream
resource "aws_iam_policy" "function_logging_policy" {
  name        = "AllowLambdaLoggingPolicy"
  description = "Policy for lambda cloudwatch logging"
  policy      = data.aws_iam_policy_document.allow_lambda_logging.json
}

// attach policy to out created lambda role
resource "aws_iam_role_policy_attachment" "lambda_logging_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_logging_policy.arn
}

// create a policy to grant lambda access to a bucket
data "aws_iam_policy_document" "allow_s3_access" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:GetObject",
      "s3:CopyObject",
      "s3:HeadObject"
    ]
    resources = [
      aws_s3_bucket.csv_txns_files.arn,
      "${aws_s3_bucket.csv_txns_files.arn}/*"
    ]
  }
}

// create a policy to allow reading from s3 bucket
resource "aws_iam_policy" "function_s3_policy" {
  name        = "AllowS3AccessPolicy"
  description = "Policy for lambda access to S3 bucket"
  policy      = data.aws_iam_policy_document.allow_s3_access.json
}

// attach policy to out created lambda role
resource "aws_iam_role_policy_attachment" "lambda_s3_policy_atachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_s3_policy.arn
}

// grant source s3 bucket permission to trigger lambda function
resource "aws_lambda_permission" "allow_terraform_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.function.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.csv_txns_files.arn
}
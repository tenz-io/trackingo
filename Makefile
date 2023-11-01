
.PHONY:
mock:
	find . -name '*_mock.go' -delete
	mockery --all --inpackage --inpackage-suffix --case underscore
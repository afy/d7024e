
#!/bin/bash

# Check if the number of iterations is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <number of times to run>"
  exit 1
fi

n=$1    # Number of times to run the test
success=0
failure=0

# Loop n times
for (( i=1; i<=n; i++ ))
do
  printf "\rRunning iteration $i..."
  
  # Run the go test -cover command
  go test -cover > /dev/null
  exit_code=$?

  # Check the exit code
  if [ $exit_code -eq 0 ]; then
    success=$((success+1))
  else
    failure=$((failure+1))
  fi
done

# Print the results
echo "Total runs: $n"
echo "Successful runs: $success"
echo "Failed runs: $failure"


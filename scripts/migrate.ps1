# Load .env file into environment variables
Get-Content .env | ForEach-Object { #Reads your .env file line by line
    if ($_ -match '^([^#][^=]+)=(.+)$') { #Checks if line is: KEY=value, ignores comments (#), splits into key and value
        Set-Item -Path "env:$($matches[1])" -Value $matches[2] #Converts .env into real environment variables
    }
}

# Get command line arguments
$command = $args[0]
$name = $args[1]

# Handle commands
switch ($command) {

    "up" {
        migrate -path migrations -database $env:DATABASE_URL up
    }

    "down" {
        $count = if ($name) { $name } else { "1" } #Determine how many migrations to rollback, default is 1

        Write-Host "Rolling back $count migration(s). Continue? [y/N]"
        $confirm = Read-Host

        if ($confirm -eq 'y') {
            migrate -path migrations -database $env:DATABASE_URL down $count
        }
    }

    "create" {
        migrate create -ext sql -dir migrations -seq $name
    }

    "force" {
        migrate -path migrations -database $env:DATABASE_URL force $name
    }
}
package rabbitConfig

import (
	"flag"
	"github.com/jroden2/commonCore/utils"
	"github.com/rs/zerolog"
	"os"
	"path"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var runningLocally = flag.Bool("local", false, "Set for running locally. Allows for pretty logging")
var logger *zerolog.Logger

func setup() {
	tLogger := utils.InitialiseLogger()
	logger = &tLogger

	envVariablePath := path.Join("/", "vault", "secrets", "vault-secrets.properties")

	flag.Parse()
	if *runningLocally {
		envVariablePath = path.Join("../../", ".env")
	}

	err := godotenv.Load(envVariablePath)
	if err != nil {
		logger.Info().Msg("INFO: unable to load .env file")
	}

	if *runningLocally {
		_ = os.Setenv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
		_ = os.Setenv("DEPLOYMENT", "local")
	}
}

func TestAssertSetup(t *testing.T) {
	setup()
	if *runningLocally {
		val := os.Getenv("DEPLOYMENT")
		assert.Equal(t, "local", val)
	}
}

func TestNewRabbitMQ(t *testing.T) {
	setup()

	t.Run("With Logger", func(t *testing.T) {
		got := NewRabbitMQ(logger)
		assert.NotNil(t, got)
	})
	t.Run("Without Logger", func(t *testing.T) {
		got := NewRabbitMQ(nil)
		assert.NotNil(t, got)
	})
}

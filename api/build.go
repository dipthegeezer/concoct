package api

import (
	packer "github.com/mitchellh/packer/packer"
	cmdcommon "github.com/mitchellh/packer/common/command"
	"fmt"
	"log"
	"sync"
	"os"
	"os/signal"
	"strconv"
	"bytes"
)

type Build struct {
	Meta
}

func (b Build) Run(data *packer.Template) int{
	buildOptions := new(cmdcommon.BuildOptions)
	env, err := b.Meta.Environment()
	if err != nil {
		b.Meta.EnvConfig.Ui.Error(fmt.Sprintf("Error initializing environment: %s", err))
		return 1
	}

	env.Ui().Say("Reached this point.")

	// The component finder for our builds
	components := &packer.ComponentFinder{
		Builder:       env.Builder,
		Hook:          env.Hook,
		PostProcessor: env.PostProcessor,
		Provisioner:   env.Provisioner,
	}

	// Go through each builder and compile the builds that we care about
	builds, err := buildOptions.Builds(data, components)
	if err != nil {
		env.Ui().Error(err.Error())
		return 1
	}
	env.Ui().Say("here")

	// Set the debug and force mode and prepare all the builds
	for _, build := range builds {
		log.Printf("Preparing build: %s", build.Name())
		env.Ui().Say(fmt.Sprintf("Preparing build: %s", build.Name()))
		//b.SetDebug(cfgDebug)
		//b.SetForce(cfgForce)

		warnings, err := build.Prepare()
		if err != nil {
			env.Ui().Error(err.Error())
			return 1
		}
		if len(warnings) > 0 {
			//ui := buildUis[build.Name()]
			env.Ui().Say(fmt.Sprintf("Warnings for build '%s':\n", build.Name()))
			for _, warning := range warnings {
				env.Ui().Say(fmt.Sprintf("* %s", warning))
			}
			env.Ui().Say("")
		}
	}

	// Run all the builds in parallel and wait for them to complete
	var interruptWg, wg sync.WaitGroup
	interrupted := false
	artifacts := make(map[string][]packer.Artifact)
	errors := make(map[string]error)
	for _, b := range builds {
		// Increment the waitgroup so we wait for this item to finish properly
		wg.Add(1)

		// Handle interrupts for this build
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)
		go func(b packer.Build) {
			<-sigCh
			interruptWg.Add(1)
			defer interruptWg.Done()
			interrupted = true

			log.Printf("Stopping build: %s", b.Name())
			b.Cancel()
			log.Printf("Build cancelled: %s", b.Name())
		}(b)

		// Run the build in a goroutine
		go func(b packer.Build) {
			defer wg.Done()

			name := b.Name()
			log.Printf("Starting build run: %s", name)
			ui := env.Ui()
			runArtifacts, err := b.Run(ui, env.Cache())

			if err != nil {
				ui.Error(fmt.Sprintf("Build '%s' errored: %s", name, err))
				errors[name] = err
			} else {
				ui.Say(fmt.Sprintf("Build '%s' finished.", name))
				artifacts[name] = runArtifacts
			}
		}(b)


		if interrupted {
			log.Println("Interrupted, not going to start any more builds.")
			break
		}
	}

	// Wait for both the builds to complete and the interrupt handler,
	// if it is interrupted.
	log.Printf("Waiting on builds to complete...")
	wg.Wait()

	log.Printf("Builds completed. Waiting on interrupt barrier...")
	interruptWg.Wait()

	if interrupted {
		env.Ui().Say("Cleanly cancelled builds after being interrupted.")
		return 1
	}

	if len(errors) > 0 {
		env.Ui().Machine("error-count", strconv.FormatInt(int64(len(errors)), 10))

		env.Ui().Error("\n==> Some builds didn't complete successfully and had errors:")
		for name, err := range errors {
			// Create a UI for the machine readable stuff to be targetted
			ui := &packer.TargettedUi{
				Target: name,
				Ui:     env.Ui(),
			}

			ui.Machine("error", err.Error())

			env.Ui().Error(fmt.Sprintf("--> %s: %s", name, err))
		}
	}

	if len(artifacts) > 0 {
		env.Ui().Say("\n==> Builds finished. The artifacts of successful builds are:")
		for name, buildArtifacts := range artifacts {
			// Create a UI for the machine readable stuff to be targetted
			ui := &packer.TargettedUi{
				Target: name,
				Ui:     env.Ui(),
			}

			// Machine-readable helpful
			ui.Machine("artifact-count", strconv.FormatInt(int64(len(buildArtifacts)), 10))

			for i, artifact := range buildArtifacts {
				var message bytes.Buffer
				fmt.Fprintf(&message, "--> %s: ", name)

				if artifact != nil {
					fmt.Fprintf(&message, artifact.String())
				} else {
					fmt.Fprint(&message, "<nothing>")
				}

				iStr := strconv.FormatInt(int64(i), 10)
				if artifact != nil {
					ui.Machine("artifact", iStr, "builder-id", artifact.BuilderId())
					ui.Machine("artifact", iStr, "id", artifact.Id())
					ui.Machine("artifact", iStr, "string", artifact.String())

					files := artifact.Files()
					ui.Machine("artifact",
					iStr,
					"files-count", strconv.FormatInt(int64(len(files)), 10))
					for fi, file := range files {
						fiStr := strconv.FormatInt(int64(fi), 10)
						ui.Machine("artifact", iStr, "file", fiStr, file)
					}
				} else {
					ui.Machine("artifact", iStr, "nil")
				}

				ui.Machine("artifact", iStr, "end")
				env.Ui().Say(message.String())
			}
		}
	} else {
		env.Ui().Say("\n==> Builds finished but no artifacts were created.")
	}

	if len(errors) > 0 {
		// If any errors occurred, exit with a non-zero exit status
		return 1
	}

	return 0
}
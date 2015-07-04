package chains

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
)

var erisDir string = path.Join(os.TempDir(), "eris")
var chainName string = "testchain"
var hash string

func TestMain(m *testing.M) {
	if err := testsInit(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	exitCode := m.Run()

	var e1, e2, e3 error
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		e1 = data.RmDataRaw(chainName, 1)
		if e1 != nil {
			fmt.Println(e1)
		}
	}

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		e3 = testsTearDown()
		if e3 != nil {
			fmt.Println(e3)
		}
	}

	if e1 != nil || e2 != nil || e3 != nil {
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func TestKnownChainRaw(t *testing.T) {
	k := services.ListKnownRaw()
	if len(k) < 2 {
		fmt.Printf("Less than two service definitions found. Something is wrong.\n")
		t.Fail()
		testsTearDown()
		os.Exit(1)
	}

	if k[0] != "erisdb" {
		fmt.Printf("Could not find erisdb service definition.\n")
		t.Fail()
		testsTearDown()
		os.Exit(1)
	}
}

func TestNewChainRaw(t *testing.T) {
	genFile := path.Join(common.BlockchainsPath, "genesis", "default.json")
	e := NewChainRaw("erisdb", chainName, genFile, "", "", 1) // configFile and dir are not needed for the tests.
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
}

func TestLoadChainDefinition(t *testing.T) {
	var e error
	chn, e := LoadChainDefinition(chainName, 1)
	if e != nil {
		fmt.Println(e)
		t.FailNow()
	}

	if chn.Service.Name != chainName {
		fmt.Printf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s\n", chainName, chn.Service.Name)
		t.FailNow()
	}

	if !chn.Operations.DataContainer {
		fmt.Println(chn.Operations)
		fmt.Printf("FAILURE: data_container not properly read on LOAD.\n")
		t.FailNow()
	}

	if chn.Operations.DataContainerName == "" {
		fmt.Printf("FAILURE: data_container_name not set.\n")
		t.Fail()
	}
}

func TestStartChainRaw(t *testing.T) {
	e := StartChainRaw(chainName, 1)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, chainName, true, true)
}

func TestLogsChainRaw(t *testing.T) {
	e := LogsChainRaw(chainName, false, 1)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
}

func TestExecChainRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		fmt.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
		return
	}
	cmd := strings.Fields("ls -la /home/eris/.eris/blockchains")
	e := ExecChainRaw(chainName, cmd, false, 1)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
}

func TestUpdateChainRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		fmt.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	e := UpdateChainRaw(chainName, true, 1)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, chainName, true, true)
}

func TestRenameChainRaw(t *testing.T) {
	e := RenameChainRaw(chainName, "niahctset")
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, "niahctset", true, true)

	e = RenameChainRaw("niahctset", chainName)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, chainName, true, true)
}

func TestKillChainRaw(t *testing.T) {
	testRunAndExist(t, chainName, true, true)

	e := KillChainRaw(chainName, false, false, 1)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, chainName, true, false)
}

func TestRmChainRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		fmt.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	e := RmChainRaw(chainName, false, false, 1)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	testRunAndExist(t, chainName, false, false)
}

func testsInit() error {
	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	util.ChangeErisDir(erisDir)

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	if err := util.Initialize(false, false); err != nil {
		return fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n")
	}

	// init dockerClient
	util.DockerConnect(false)

	// make sure erisdb not running
	for _, r := range services.ListRunningRaw() {
		if r == "erisdb" {
			return fmt.Errorf("ERISDB service is running. Please stop it with eris services stop erisdb.")
		}
	}

	// make sure erisdb container does not exist
	for _, r := range services.ListExistingRaw() {
		if r == "erisdb" {
			return fmt.Errorf("ERISDB service exists. Please remove it with eris services rm erisdb.")
		}
	}

	return nil
}

func testsTearDown() error {
	if e := os.RemoveAll(erisDir); e != nil {
		return e
	}

	return nil
}

func testRunAndExist(t *testing.T, chainName string, toExist, toRun bool) {
	var exist, run bool
	for _, r := range ListExistingRaw() {
		if r == chainName {
			exist = true
		}
	}
	for _, r := range ListRunningRaw() {
		if r == chainName {
			run = true
		}
	}

	if toRun != run {
		if toRun {
			fmt.Printf("Could not find a running instance of %s\n", chainName)
			t.Fail()
		} else {
			fmt.Printf("Found a running instance of %s when I shouldn't have\n", chainName)
			t.Fail()
		}
	}

	if toExist != exist {
		if toExist {
			fmt.Printf("Could not find an existing instance of %s\n", chainName)
			t.Fail()
		} else {
			fmt.Printf("Found an existing instance of %s when I shouldn't have\n", chainName)
			t.Fail()
		}
	}
}
package provision

import (
	"errors"
	"fmt"
	"strings"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

type Provisioner struct {
	Config       entity.Config
	Packager     packager
	provisioners provisioners
}

type provisionFn struct {
	name string
	fn   func() error
}

type provisioners struct {
	provMap map[string]func() error
	order   []string
}

func (p provisioners) apply() error {
	var provErrs []error
	for _, fnName := range p.order {
		fn := p.provMap[fnName]
		err := fn()
		provErrs = append(provErrs, err)
	}

	var uniqErrs []error
	seenErr := make(map[string]error)
	for _, err := range provErrs {
		if err == nil {
			continue
		}
		msg := err.Error()
		_, ok := seenErr[msg]
		if !ok {
			seenErr[msg] = err
			uniqErrs = append(uniqErrs, err)
		}
	}

	return errors.Join(uniqErrs...)
}

func getOrderedProvisioners(provFns []provisionFn) []string {
	var names []string
	for _, fn := range provFns {
		names = append(names, fn.name)
	}
	return names
}

func newProvisioners(allProvisioners []provisionFn, filter []string) (provisioners, error) {
	provMap := make(map[string]func() error)
	var order []string
	var hasFilter = len(filter) > 0

	for _, prov := range allProvisioners {
		provMap[prov.name] = prov.fn
		if !hasFilter {
			order = append(order, prov.name)
		}
	}

	if hasFilter {
		for _, fnName := range filter {
			_, ok := provMap[fnName]
			if !ok {
				ordered := getOrderedProvisioners(allProvisioners)
				msg := fmt.Sprintf("available provisioners (in order of execution): %s", strings.Join(
					ordered, " "))
				return provisioners{}, fmt.Errorf("%s is not a provisioning function, %s", fnName, msg)
			}

			order = append(order, fnName)
		}
	}

	return provisioners{
		provMap: provMap,
		order:   order,
	}, nil
}

func NewProvisioner(cfg entity.Config, filter []string) (Provisioner, error) {
	pkgr, err := newPackager()
	if err != nil {
		return Provisioner{}, err
	}

	p := Provisioner{Config: cfg, Packager: pkgr}

	all := []provisionFn{
		{"pre", p.runPreflightTasks},
		{"repo", p.AddOSRepos},
		{"release", p.ensureReleases},
		{"package", p.installPackages},
		{"host", p.acceptHostKeys},
		{"github", p.githubUserKey},
		{"go", p.goInstall},
		{"python", p.pythonInstall},
		{"rust", p.rustInstall},
		{"clone", p.sshClone},
		{"task", p.runTasks},
		{"template", p.applyTemplates},
		{"service", p.initServices},
		{"dir", p.ensureDirs},
		{"line", p.ensureLines},
		{"archive", p.extractArchive},
		{"flatpak", p.flatpakInstall},
		{"snap", p.snapInstall},
		{"group", p.userInGroup},
		{"post", p.runPostFlightTasks},
	}

	provs, err := newProvisioners(all, filter)
	if err != nil {
		return p, err
	}

	p.provisioners = provs
	return p, nil
}

func (p Provisioner) Apply() error {
	err := evalFacts(p.Config)
	if err != nil {
		return err
	}

	return p.provisioners.apply()
}

func (p Provisioner) AddOSRepos() error {
	internal.Logger.Info().Msg("Adding OS repos")

	return addRepos(p.Config)
}

func (p Provisioner) ensureReleases() error {
	internal.Logger.Info().Msg("Downloading releases")

	if p.Config.Settings.ReleaseDir == "" {
		return errors.New("empty release directory")
	}

	return ensureReleases(p.Config)
}

func (p Provisioner) runPreflightTasks() error {
	internal.Logger.Info().Msg("Running preflight tasks")

	return runTasks(p.Config, p.Config.PreflightTasks)
}

func (p Provisioner) runPostFlightTasks() error {
	internal.Logger.Info().Msg("Running postflight tasks")

	return runTasks(p.Config, p.Config.PostflightTasks)
}

func (p Provisioner) installPackages() error {
	internal.Logger.Info().Msg("Installing/removing packages")

	return installPackages(p)
}

func (p Provisioner) rustInstall() error {
	internal.Logger.Info().Msg("Installing Rust packages")

	return cargoInstallPkgs(p.Config)
}

func (p Provisioner) githubUserKey() error {
	internal.Logger.Info().Msg("Adding GitHub user keys")

	return addGithubUserKeys(p.Config)
}

func (p Provisioner) goInstall() error {
	internal.Logger.Info().Msg("Installing Go packages")

	return goInstallPkgs(p.Config)
}

func (p Provisioner) acceptHostKeys() error {
	internal.Logger.Info().Msg("Adding known hosts")

	return acceptHostKeys(p.Config)
}

func (p Provisioner) initServices() error {
	internal.Logger.Info().Msg("Initializing services")

	return initServices(p.Config)
}

func (p Provisioner) pythonInstall() error {
	internal.Logger.Info().Msg("Installing Python packages")

	return pythonInstallPkgs(p.Config)
}

func (p Provisioner) runTasks() error {
	internal.Logger.Info().Msg("Running tasks")

	return runTasks(p.Config, p.Config.Tasks)
}

func (p Provisioner) applyTemplates() error {
	internal.Logger.Info().Msg("Applying templates")

	return applyTemplates(p.Config)
}

func (p Provisioner) ensureDirs() error {
	internal.Logger.Info().Msg("Creating desired dirs")

	return ensureDirs(p.Config)
}

func (p Provisioner) ensureLines() error {
	internal.Logger.Info().Msg("Ensuring lines in files")

	return ensureLines(p.Config)
}

func (p Provisioner) extractArchive() error {
	internal.Logger.Info().Msg("Extracting archives")

	return extractArchives(p.Config)
}

func (p Provisioner) flatpakInstall() error {
	internal.Logger.Info().Msg("Installing Flatpak packages")

	_, err := common.Which("flatpak")
	if err != nil {
		internal.Logger.Trace().Msg("Flatpak is not installed")
		return nil
	}

	return flatpakInstall(p.Config)
}

func (p Provisioner) snapInstall() error {
	internal.Logger.Info().Msg("Installing snap packages")

	return snapInstall(p.Config)
}

func (p Provisioner) sshClone() error {
	internal.Logger.Info().Msg("Cloning repos via SSH")

	return sshClone(p.Config)
}

func (p Provisioner) userInGroup() error {
	internal.Logger.Info().Msg("Ensuring user is in desired groups")

	return userInGroup(p.Config)
}

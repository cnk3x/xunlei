# Spk Script Environment Variables

Several variables are exported by Package Center and can be used in the scripts. Descriptions of the variables are given as below:

- `SYNOPKG_PKGNAME`: Package identify which is defined in INFO.
- `SYNOPKG_PKGVER`: Package version which is defined in INFO. The value will be new version of package when it is upgrading.
- `SYNOPKG_PKGDEST`: Target directory where the package is stored.
- `SYNOPKG_PKGDEST_VOL`: Target volume where the package is stored.
- `SYNOPKG_PKGPORT`: adminport port which is defined in INFO. This port will be occupied by this package with its management interface.
- `SYNOPKG_PKGINST_TEMP_DIR`: The temporary directory where the package are extracted when installing or upgrading it.
- `SYNOPKG_TEMP_LOGFILE`: A temporary file path for a script to log information or error messages.
- `SYNOPKG_TEMP_UPGRADE_FOLDER`: The temporary directory when the package is upgrading. You can move the files from the previous version of the package to it in preupgrade script and move them back in postupgrade.
- `SYNOPKG_DSM_LANGUAGE`: End user's DSM language.
- `SYNOPKG_DSM_VERSION_MAJOR`: End user’s major number of DSM version which is formatted as [DSM major number].[DSM minor number]-[DSM build number].
- `SYNOPKG_DSM_VERSION_MINOR`: End user’s minor number of DSM version which is formatted as [DSM major number].[DSM minor number]-[DSM build number].
- `SYNOPKG_DSM_VERSION_BUILD`: End user’s DSM build number of DSM version which is formatted as [DSM major number].[DSM minor number]-[DSM build number].
- `SYNOPKG_DSM_ARCH`: End user’s DSM CPU architecture. Please refer Appendix A: Platform and Arch Value Mapping Table to more information
- `SYNOPKG_PKG_STATUS`: Package status presented by these values: INSTALL, UPGRADE, UNINSTALL, START, STOP or empty.
  - `INSTALL` will be set as the status value in the preinst and postinst scripts while the package is installing. If the user chooses to “start after installation” at the last step of the installation wizard, the value will be set to INSTALL in the start-stop-status script when the package is started.
  - `UPGRADE` will be set as the status value in the preupgrade, preuninst, postunist, preinst, postinst and postupgrade scripts sequentially while the package is upgrading. If the package has already started before upgrade, the value will be set to UPGRADE in the start-stop-status script when the package is started or stopped.
  - `UNINSTALL` will be set as the status value in the preuninst and postunist scripts while the package is un-installing. If the package has already started before un-installation, the value will be set to UNINSTALL in the start-stop-status script when the package is stopped.
  - If the user starts or stops a package in the Package Center, START or STOP will be set as the status value in the start-stop-status script.
  - When the NAS is booting up or shutting down, its status value will be empty.
- `SYNOPKG_OLD_PKGVER`: Old package version which is defined in INFO during upgrading.
- `SYNOPKG_TEMP_SPKFILE`: The location of package spk file is temporarily stored in DS when the package is installing/upgrading.
- `SYNOPKG_USERNAME`: The user name who installs, upgrades, uninstalls, starts or stops the package. If the value is empty, the action is triggered by DSM, not by the end user.
- `SYNOPKG_PKG_PROGRESS_PATH`: A temporary file path for a script to showing the progress in installing and upgrading a package.
  Note:
  - The progress value is between 0 and 1.
  - Example:

    ```shell
    flock -x "$SYNOPKG_PKG_PROGRESS_PATH" -c echo 0.80 >     "$SYNOPKG_PKG_PROGRESS_PATH"
    ```

## referer

<https://help.synology.com/developer-guide/synology_package/script_env_var.html>

<https://help.synology.com/developer-guide/appendix/platarchs.html>

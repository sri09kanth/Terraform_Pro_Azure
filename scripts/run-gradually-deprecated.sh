#!/usr/bin/env bash

function runGraduallyDeprecatedFunctions {
  echo "==> Checking for use of gradually deprecated functions..."

  IFS=$'\n' read -r -d '' -a flist < <(git diff --diff-filter=AMRC origin/main --name-only)

  for f in "${flist[@]}"; do
    # require resources to be imported is now hard-coded on - but only checking for additions
    grep -H -n "features\.ShouldResourcesBeImported" "$f" && {
        echo "The Feature Flag for 'ShouldResourcesBeImported' will be deprecated in the future"
        echo "and shouldn't be used in new resources - please remove new usages of the"
        echo "'ShouldResourcesBeImported' function from these changes - since this is now enabled"
        echo "by default."
        echo ""
        echo "In the future this function will be marked as Deprecated - however it's not for"
        echo "the moment to not conflict with open Pull Requests."
        exit 1
    }

    # using Resource ID Formatters/Parsers
    grep -H -n "d\.SetId(\\*" "$f" && {
        echo "Due to the Azure API returning the Resource ID's inconsistently - Terraform"
        echo "now manages it's own Resource ID's, all new resources should use a generated"
        echo "Resource ID Formatter and Parser."
        echo ""
        echo "A Resource ID Formatter and Parser can be generated by adding a 'resourceids.go'"
        echo "file to the service package (for example"
        echo "./internal/services/myservice/resourceids.go) - with the line:"
        echo
        echo "//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Server"
        echo "-id={the value of the Resource ID}"
        echo ""
        echo "At which point running 'make generate' will generate a Resource ID Formatter, Parser"
        echo "and Validator."
        echo ""
        echo "This allows a Resource ID to be defined in-line via:"
        echo "  > subscriptionId := meta.(*clients.Client).Account.SubscriptionId"
        echo "  > id := parse.NewMyResourceId(subscriptionId, resourceGroup, name)"
        echo ""
        echo "This means that the 'SetID' function can change from:"
        echo "  > d.SetId(\"*read.ID\")"
        echo "to:"
        echo "  > d.SetId(id.ID())"
        echo ""
        echo "In addition when parsing the Resource ID during a Read, Update or Delete method"
        echo "the generated Resource ID Parser can be used via:"
        echo "  > id, err := parse.MyResourceID(d.Id())"
        echo "  > if err != nil {"
        echo "  >   return err"
        echo "  > }"
        echo ""
        echo "New Resources should be using Resource ID Formatters/Parsers by default"
        echo "however existing (unmodified) resources can continue to use the Azure ID"
        echo "for the moment - but over time these will be switched across."
        exit 1
    }

    # check for new combined CreateUpdate methods
    line=$(grep -H -n "Create:.*CreateUpdate," "$f" -m1)
    if [ "$line" != "" ];
    then
      git diff --diff-filter=AMC origin/main -U0 "$f" | grep -q "+.*Create:.*CreateUpdate," && {
        echo "$line"
        echo "New Resources should no longer use combined CreateUpdate methods, please"
        echo "split these into two separate Create and Update methods."
        echo ""
        echo "Existing resources can continue to use combined CreateUpdate methods"
        echo "for the moment - but over time these will be split into separate Create and"
        echo "Update methods."
        exit 1
      }
    fi

    # check for d.Get inside Delete
    deleteFuncName=$(grep -o "Delete: .*," "$f" -m1 | grep -o " .*Delete"| tr -d " ")
    if [ "$deleteFuncName" != "" ];
    then
      deleteMethod=$(cat -n $f | sed -n -e "/func $deleteFuncName.*$/,/[[:digit:]]*\treturn nil/{ /func $deleteFuncName$/d; /[[:digit:]]*\treturn nil/d; p; }")
      foundGet=$(echo "$deleteMethod" | grep "d\.Get(.*)" -m1)
      if [ "$foundGet" != "" ];
      then
        echo "$f $foundGet"
        echo "Please do not use 'd.Get' within the Delete function as this does not work as expected in Delete"
        exit 1
      fi
    else
      # check for Get in typed resource
      deleteFuncName=" Delete() sdk.ResourceFunc "
      deleteMethod=$(cat -n $f | sed -n -e "/$deleteFuncName.*$/,/[[:digit:]]*\t\t\treturn nil/{ /$deleteFuncName.*$/d; /[[:digit:]]*\t\t\treturn nil/d; p; }")
      foundGet=$(echo "$deleteMethod" | grep "metadata.ResourceData.Get" -m1)
      if [ "$foundGet" != "" ];
      then
        echo "$f $foundGet"
        echo "Please do not use 'metadata.ResourceData.Get' within the Delete function as this does not work as expected in Delete"
        exit 1
      fi
    fi
  done
}

function checkForUnclearErrorMessages {
  result=$(grep -R "invalid format of " ./internal)
  if [ "$result" != "" ];
  then
    echo "The error messages in these files aren't descriptive, please add more"
    echo "context for how users can fix these. For example, changing \"invalid"
    echo "format of 'foo'\" can clearer as \"'foo' must start with letter, can"
    echo "contain both letters and numbers and must end with a letter\"."
    echo ""
    echo "$result"
    exit 1
  fi
}

function runDeprecatedFunctions {
  echo "==> Checking for use of deprecated functions..."
  result=$(grep -Ril "d.setid(\"\")" ./internal/services/**/data_source_*.go)
  if [ "$result" != "" ];
  then
    echo "Data Sources should return an error when a resource cannot be found rather than"
    echo "setting an empty ID (by calling 'd.SetId("")'."
    echo ""
    echo "Please remove the references to 'd.SetId("") from the Data Sources listed below"
    echo "and raise an error instead:"
    echo ""
    exit 1
  fi
}

function main {
  if [ "$GITHUB_ACTIONS_STAGE" == "UNIT_TESTS" ];
  then
    echo "Skipping - the Gradually Deprecated check is separate in Github Actions"
    echo "so this can be skipped as a part of the build process."
    exit 0
  fi

  runGraduallyDeprecatedFunctions
  runDeprecatedFunctions
  checkForUnclearErrorMessages
}

main

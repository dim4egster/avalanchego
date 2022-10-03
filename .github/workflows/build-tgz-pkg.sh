PKG_ROOT=/tmp
VERSION=$TAG
AVALANCHE_ROOT=$PKG_ROOT/qmallgo-$VERSION

mkdir -p $AVALANCHE_ROOT

OK=`cp ./build/qmallgo $AVALANCHE_ROOT`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi
OK=`cp -r ./build/plugins $AVALANCHE_ROOT`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi


echo "Build tgz package..."
cd $PKG_ROOT
echo "Version: $VERSION"
tar -czvf "qmallgo-linux-$ARCH-$VERSION.tar.gz" qmallgo-$VERSION
aws s3 cp qmallgo-linux-$ARCH-$VERSION.tar.gz s3://$BUCKET/linux/binaries/ubuntu/$RELEASE/$ARCH/
rm -rf $PKG_ROOT/qmallgo*

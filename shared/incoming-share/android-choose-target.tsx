import * as React from 'react'
import * as Container from '../util/container'
import * as RouteTreeGen from '../actions/route-tree-gen'
import * as FsGen from '../actions/fs-gen'
import * as FsTypes from '../constants/types/fs'
import * as RPCTypes from '../constants/types/rpc-gen'
import ChooseTarget from './choose-target'

const AndroidChooseTarget = () => {
  const dispatch = Container.useDispatch()
  const onBack = () => dispatch(RouteTreeGen.createNavigateUp())
  const share = Container.useSelector(state => state.config.androidShare)
  const onKBFS =
    share?.type === RPCTypes.IncomingShareType.file
      ? () => {
          dispatch(FsGen.createSetIncomingShareSource({source: FsTypes.stringToLocalPath(share.url)}))
          dispatch(
            FsGen.createShowIncomingShare({initialDestinationParentPath: FsTypes.stringToPath('/keybase')})
          )
        }
      : // Disable sharing text into KBFS on Android for now.
        undefined

  const item = share
    ? share.type === RPCTypes.IncomingShareType.file
      ? {
          payloadPath: share.url,
          type: RPCTypes.IncomingShareType.file,
        }
      : {
          content: share.text,
          type: RPCTypes.IncomingShareType.text,
        }
    : undefined
  const onChat = item
    ? () =>
        dispatch(
          RouteTreeGen.createNavigateAppend({
            path: [
              {
                props: {
                  incomingShareItems: [item],
                },
                selected: 'sendAttachmentToChat',
              },
            ],
          })
        )
    : undefined

  return item ? (
    <ChooseTarget
      items={[
        {
          filename: item.payloadPath ? FsTypes.getLocalPathName(item.payloadPath) : '',
          shareType: item.type,
        },
      ]}
      onBack={onBack}
      onChat={onChat}
      onKBFS={onKBFS}
    />
  ) : null
}

export default AndroidChooseTarget
